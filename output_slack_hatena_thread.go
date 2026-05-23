package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/slack-go/slack"
)

// HatenaEntryState は、Slackへの投稿状態を管理するための構造体です
type HatenaEntryState struct {
	LastCommentTimestamp string `json:"last_comment_timestamp"`
	SlackThreadTimestamp string `json:"slack_thread_timestamp"`
}

// State は、URLをキーとした状態のマップです
type State map[string]HatenaEntryState

const (
	stateFilePath = "hatena_state.json"
)

type hatenaOutputOptions struct {
	Token                 string
	Channel               string
	DryRun                bool
	SkipChannelValidation bool
	DryRunWriter          io.Writer
	API                   slackClient
	StateBackend          string
	StatePath             string
}

// OutputHatenaToSlack は、フィルタリングされたフィードを処理し、Slackに通知します
func OutputHatenaToSlack(items []*FilteredItem) {
	if err := OutputHatenaToSlackWithOptions(items, hatenaOutputOptions{}); err != nil {
		log.Fatal(err)
	}
}

func OutputHatenaToSlackWithOptions(items []*FilteredItem, options hatenaOutputOptions) error {
	slackChannelID := slackChannelFromOptions(options.Channel)
	store, err := newHatenaStateStore(options)
	if err != nil {
		return err
	}
	if options.DryRun {
		return outputHatenaToSlackDryRun(items, slackChannelID, store, options.DryRunWriter)
	}

	token := options.Token
	if token == "" {
		token = slackTokenFromEnv()
	}
	if token == "" {
		return errors.New("Slack tokenを設定してください。FEED2CLI_SLACK_TOKENまたはXOXBを使用できます。")
	}
	if slackChannelID == "" {
		return errors.New("Slack channelを設定してください。-slack-channel、FEED2CLI_SLACK_CHANNEL、SLACK_CHANNELのいずれかを使用できます。")
	}

	api := options.API
	if api == nil {
		api = slack.New(token)
	}
	resolvedChannel, err := prepareSlackDestination(api, slackChannelID, !options.SkipChannelValidation)
	if err != nil {
		return err
	}
	return outputHatenaToSlack(items, api, resolvedChannel, store, time.Sleep)
}

type hatenaStateStore interface {
	Load() (State, error)
	Save(State) error
}

type fileHatenaStateStore struct {
	path string
}

func outputHatenaToSlack(items []*FilteredItem, api slackPoster, slackChannelID string, store hatenaStateStore, sleep func(time.Duration)) error {
	if api == nil {
		return errors.New("slack poster is nil")
	}
	if store == nil {
		return errors.New("state store is nil")
	}
	if sleep == nil {
		sleep = time.Sleep
	}

	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("状態ファイルの読み込みに失敗しました: %w", err)
	}

	stateChanged := false
	failed := 0

	for _, item := range items {
		if item == nil || item.Item == nil {
			continue
		}
		entryURL := item.Link
		entryKey := itemDedupKey(item.Item)
		log.Printf("処理中のエントリ: %s", entryURL)

		entryState, exists := state[entryKey]
		if !exists && entryKey != entryURL {
			if rawState, rawExists := state[entryURL]; rawExists {
				entryState = rawState
				state[entryKey] = rawState
				delete(state, entryURL)
				exists = true
				stateChanged = true
			}
		}
		if !exists {
			attachment := slack.Attachment{
				Title:     item.Title,
				TitleLink: entryURL,
				Text:      item.Description,
			}
			channelID, timestamp, err := api.PostMessage(
				slackChannelID,
				slack.MsgOptionAttachments(attachment),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				log.Printf("Slackへのメッセージ投稿に失敗しました: %v", err)
				failed++
				continue
			}
			log.Printf("新規エントリをSlackに投稿しました: Channel=%s, Timestamp=%s", channelID, timestamp)

			entryState = HatenaEntryState{
				SlackThreadTimestamp: timestamp,
				LastCommentTimestamp: "1970-01-01T00:00:00Z",
			}
			state[entryKey] = entryState
			stateChanged = true
		}

		// FilteredItemからコメント情報を取得
		comments := item.HatenaBookmarkComments

		// 差分コメントを投稿
		lastPostTime, _ := time.Parse(time.RFC3339, entryState.LastCommentTimestamp)
		latestPostedCommentTime := lastPostTime
		for _, comment := range comments {
			commentTime, err := time.Parse("2006/01/02 15:04", comment.Timestamp)
			if err != nil {
				log.Printf("コメントのタイムスタンプのパースに失敗しました: %v", err)
				continue
			}

			if commentTime.After(lastPostTime) {
				commentText := fmt.Sprintf("*%s* さん: %s", comment.User, comment.Comment)
				_, _, err := api.PostMessage(
					slackChannelID,
					slack.MsgOptionText(commentText, false),
					slack.MsgOptionTS(entryState.SlackThreadTimestamp),
					slack.MsgOptionAsUser(true),
				)
				if err != nil {
					log.Printf("Slackスレッドへの投稿に失敗しました: %v", err)
					failed++
				} else {
					log.Printf("新規コメントをスレッドに投稿しました: %s", commentText)
					if commentTime.After(latestPostedCommentTime) {
						latestPostedCommentTime = commentTime
					}
					stateChanged = true
				}
				// Slackのレートリミットを回避するために1.5秒待機
				sleep(1500 * time.Millisecond)
			}
		}

		// 状態を最新のコメント時刻で更新
		if latestPostedCommentTime.After(lastPostTime) {
			updatedState := state[entryKey]
			updatedState.LastCommentTimestamp = latestPostedCommentTime.Format(time.RFC3339)
			state[entryKey] = updatedState
			stateChanged = true
		}
	}

	if stateChanged {
		if err := store.Save(state); err != nil {
			return fmt.Errorf("状態ファイルの保存に失敗しました: %w", err)
		}
		log.Println("状態ファイルを更新しました。")
	} else {
		log.Println("新規コメントはありませんでした。")
	}
	if failed > 0 {
		return fmt.Errorf("Slackへの投稿に%d件失敗しました", failed)
	}
	return nil
}

func outputHatenaToSlackDryRun(items []*FilteredItem, channel string, store hatenaStateStore, w io.Writer) error {
	if store == nil {
		return errors.New("state store is nil")
	}
	if w == nil {
		w = io.Discard
	}
	if channel == "" {
		channel = "(unset)"
	}
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("状態ファイルの読み込みに失敗しました: %w", err)
	}

	parentPosts := 0
	commentPosts := 0
	for _, item := range items {
		if item == nil || item.Item == nil {
			continue
		}
		entryURL := item.Link
		entryKey := itemDedupKey(item.Item)
		entryState, exists := state[entryKey]
		if !exists && entryKey != entryURL {
			entryState, exists = state[entryURL]
		}
		if !exists {
			parentPosts++
			entryState = HatenaEntryState{LastCommentTimestamp: "1970-01-01T00:00:00Z"}
		}

		lastPostTime, _ := time.Parse(time.RFC3339, entryState.LastCommentTimestamp)
		newComments := 0
		for _, comment := range item.HatenaBookmarkComments {
			commentTime, err := time.Parse("2006/01/02 15:04", comment.Timestamp)
			if err == nil && commentTime.After(lastPostTime) {
				newComments++
			}
		}
		commentPosts += newComments
		fmt.Fprintf(w, "hatena dry-run: title=%q url=%q parent_post=%t new_comments=%d\n", item.Title, entryURL, !exists, newComments)
	}
	fmt.Fprintf(w, "hatena dry-run summary: channel=%s parent_posts=%d comment_posts=%d\n", channel, parentPosts, commentPosts)
	return nil
}

func loadState() (State, error) {
	return (fileHatenaStateStore{path: stateFilePath}).Load()
}

func saveState(s State) error {
	return (fileHatenaStateStore{path: stateFilePath}).Save(s)
}

func (store fileHatenaStateStore) Load() (State, error) {
	data, err := os.ReadFile(store.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(State), nil
		}
		return nil, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func (store fileHatenaStateStore) Save(state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(store.path, data, 0644)
}
