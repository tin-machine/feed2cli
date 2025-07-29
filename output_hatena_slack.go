package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	
	"os"
	"time"

	"github.com/mmcdole/gofeed"
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

// OutputHatenaToSlack は、はてなブックマークのフィードを処理し、Slackに通知します
func OutputHatenaToSlack(feeds []*gofeed.Feed) {
	token := os.Getenv("XOXB")
	if token == "" {
		log.Fatal("環境変数XOXBにアクセストークンを設定してください。")
	}
	slackChannelID := os.Getenv("SLACK_CHANNEL")
	if slackChannelID == "" {
		log.Fatal("環境変数SLACK_CHANNELに通知先のチャンネル名またはユーザーIDを設定してください。")
	}

	api := slack.New(token)
	state, err := loadState()
	if err != nil {
		log.Fatalf("状態ファイルの読み込みに失敗しました: %v", err)
	}

	// 変更があったかどうかを追跡
	stateChanged := false

	for _, feed := range feeds {
		for _, item := range feed.Items {
			entryURL := item.Link
			log.Printf("処理中のエントリ: %s", entryURL)

			entryState, exists := state[entryURL]
			if !exists {
				// 新規エントリの場合、Slackに親メッセージを投稿
				attachment := slack.Attachment{
					Title:     item.Title,
					TitleLink: entryURL,
					Text:      item.Description,
				}
				channelID, timestamp, err := api.PostMessage(
					slackChannelID,
					slack.MsgOptionAttachments(attachment),
					slack.MsgOptionAsUser(true), // Botとしてではなく、ユーザーとして投稿する場合
				)
				if err != nil {
					log.Printf("Slackへのメッセージ投稿に失敗しました: %v", err)
					continue
				}
				log.Printf("新規エントリをSlackに投稿しました: Channel=%s, Timestamp=%s", channelID, timestamp)

				entryState = HatenaEntryState{
					SlackThreadTimestamp: timestamp,
					LastCommentTimestamp: "1970-01-01T00:00:00Z", // 最初はUNIXエポック
				}
				state[entryURL] = entryState
				stateChanged = true
			}

			// はてなブックマークのコメントを取得
			comments, err := GetHatenaBookmarkComments(entryURL)
			if err != nil {
				log.Printf("はてなブックマークコメントの取得に失敗しました: %v", err)
				continue
			}

			// 新規コメントをフィルタリングして投稿
			newLastTimestamp := entryState.LastCommentTimestamp
			for _, comment := range comments {
				// "2023/10/27 15:04:05" のような形式をパース
				commentTime, err := time.Parse("2006/01/02 15:04:05", comment.Timestamp)
				if err != nil {
					log.Printf("コメントのタイムスタンプのパースに失敗しました: %v", err)
					continue
				}
				lastPostTime, _ := time.Parse(time.RFC3339, entryState.LastCommentTimestamp)

				if commentTime.After(lastPostTime) {
					// 新規コメントをスレッドに投稿
					commentText := fmt.Sprintf("*%s* さん: %s", comment.User, comment.Comment)
					_, _, err := api.PostMessage(
						slackChannelID,
						slack.MsgOptionText(commentText, false),
						slack.MsgOptionTS(entryState.SlackThreadTimestamp), // スレッドに返信
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Printf("Slackスレッドへの投稿に失敗しました: %v", err)
					} else {
						log.Printf("新規コメントをスレッドに投稿しました: %s", commentText)
						stateChanged = true
						// 最後に成功したコメントの時間を記録
						newLastTimestamp = commentTime.Format(time.RFC3339)
					}
				}
			}
			// 状態を更新
			if state[entryURL].LastCommentTimestamp != newLastTimestamp {
				updatedState := state[entryURL]
				updatedState.LastCommentTimestamp = newLastTimestamp
				state[entryURL] = updatedState
			}
		}
	}

	// 状態が変更されていたらファイルに保存
	if stateChanged {
		if err := saveState(state); err != nil {
			log.Fatalf("状態ファイルの保存に失敗しました: %v", err)
		}
		log.Println("状態ファイルを更新しました。")
	} else {
		log.Println("新規コメントはありませんでした。")
	}
}

// loadState は、状態管理ファイルを読み込みます
func loadState() (State, error) {
	data, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		// ファイルが存在しない場合は、空の状態で開始する
		if _, ok := err.(*os.PathError); ok {
			return make(State), nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s, nil
}

// saveState は、状態管理ファイルに現在の状態を保存します
func saveState(s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(stateFilePath, data, 0644)
}


