package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

// OutputHatenaToSlack は、フィルタリングされたフィードを処理し、Slackに通知します
func OutputHatenaToSlack(items []*FilteredItem) {
	token := os.Getenv("XOXB")
	if token == "" {
		log.Fatal("環境変数XOXBにアクセストークンを設定してください。")
	}
	slackChannelID := os.Getenv("SLACK_CHANNEL")
	if slackChannelID == "" {
		log.Fatal("環境変数SLACK_CHANNELに通知先のチャンネル名またはユーザーIDを設定してください。")
	}

	api := slack.New(token)
	store := fileHatenaStateStore{path: stateFilePath}
	if err := outputHatenaToSlack(items, api, slackChannelID, store, time.Sleep); err != nil {
		log.Fatal(err)
	}
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

	for _, item := range items {
		if item == nil || item.Item == nil {
			continue
		}
		entryURL := item.Link
		log.Printf("処理中のエントリ: %s", entryURL)

		entryState, exists := state[entryURL]
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
				continue
			}
			log.Printf("新規エントリをSlackに投稿しました: Channel=%s, Timestamp=%s", channelID, timestamp)

			entryState = HatenaEntryState{
				SlackThreadTimestamp: timestamp,
				LastCommentTimestamp: "1970-01-01T00:00:00Z",
			}
			state[entryURL] = entryState
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
			updatedState := state[entryURL]
			updatedState.LastCommentTimestamp = latestPostedCommentTime.Format(time.RFC3339)
			state[entryURL] = updatedState
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
