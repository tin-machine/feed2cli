package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

// OutputSlack は、フィードアイテムをSlackに通知します。
// フィルタによって追加された情報があれば、それもメッセージに含めます。
func OutputSlack(data interface{}) {
	token := os.Getenv("XOXB")
	if token == "" {
		log.Fatal("環境変数XOXBにアクセストークンを設定してください。")
	}
	slackChannelID := os.Getenv("SLACK_CHANNEL")
	if slackChannelID == "" {
		log.Fatal("環境変数SLACK_CHANNELに通知先のチャンネル名またはユーザーIDを設定してください。")
	}

	api := slack.New(token)

	itemsToProcess := convertToFilteredItems(data)

	for _, item := range itemsToProcess {
		// Slackメッセージ用テキストを生成
		// Descriptionは長くなる可能性があるので、Attachmentに含める
		markdownText := fmt.Sprintf("<%s|%s>", item.Link, item.Title)

		// Attachmentを生成
		attachment := formatSlackAttachment(item)

		// Slackへメッセージを送信
		_, _, err := api.PostMessage(
			slackChannelID,
			slack.MsgOptionText(markdownText, false),
			slack.MsgOptionAttachments(attachment),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			log.Printf("Slackへのメッセージ送信でエラーが発生しました: %v", err)
		}
		// レートリミット対策
		time.Sleep(1500 * time.Millisecond)
	}
}

// formatSlackAttachment は、FilteredItemからSlackのAttachmentを生成します。
func formatSlackAttachment(item *FilteredItem) slack.Attachment {
	var attachmentText strings.Builder
	
	// 元のDescriptionを追加
	attachmentText.WriteString(item.Description)
	attachmentText.WriteString("\n\n")

	// はてなブックマーク情報があれば追記
	if item.HatenaBookmarkCount != "" && item.HatenaBookmarkCount != "0" {
		attachmentText.WriteString(fmt.Sprintf("Hatena Bookmark: *%s*\n", item.HatenaBookmarkCount))
	}
	if len(item.HatenaBookmarkComments) > 0 {
		// コメントが多すぎるとメッセージが長くなるため、件数のみ表示
		attachmentText.WriteString(fmt.Sprintf("Hatena Comments: %d\n", len(item.HatenaBookmarkComments)))
	}

	return slack.Attachment{
		Text: attachmentText.String(),
	}
}

// convertToFilteredItems は、様々な型のデータを []*FilteredItem に変換します。
// これにより、後続の処理が型を意識せずに済むようになります。
func convertToFilteredItems(data interface{}) []*FilteredItem {
	if filtered, ok := data.([]*FilteredItem); ok {
		return filtered
	}

	var items []*gofeed.Item
	if feeds, ok := data.([]*gofeed.Feed); ok {
		for _, feed := range feeds {
			items = append(items, feed.Items...)
		}
	} else if feedItems, ok := data.([]*gofeed.Item); ok {
		items = feedItems
	}

	filteredItems := make([]*FilteredItem, len(items))
	for i, item := range items {
		filteredItems[i] = &FilteredItem{Item: item}
	}
	return filteredItems
}