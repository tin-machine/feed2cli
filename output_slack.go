package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

// OutputSlack は、フィードアイテムをSlackに通知します。
// フィルタによって追加された情報があれば、それもメッセージに含めます。
func OutputSlack(data interface{}) {
	if err := OutputSlackWithOptions(data, slackOutputOptions{}); err != nil {
		log.Fatal(err)
	}
}

type slackOutputOptions struct {
	Token                 string
	Channel               string
	DryRun                bool
	SkipChannelValidation bool
	DryRunWriter          io.Writer
	API                   slackClient
}

func OutputSlackWithOptions(data interface{}, options slackOutputOptions) error {
	slackChannelID := slackChannelFromOptions(options.Channel)
	if options.DryRun {
		return outputSlackDryRun(data, slackChannelID, options.DryRunWriter)
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

	return outputSlack(data, api, resolvedChannel, time.Sleep)
}

type slackPoster interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

type slackClient interface {
	slackPoster
	AuthTest() (*slack.AuthTestResponse, error)
	GetConversationInfo(channelID string, includeLocale bool) (*slack.Channel, error)
	GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error)
}

func outputSlack(data interface{}, api slackPoster, slackChannelID string, sleep func(time.Duration)) error {
	if api == nil {
		return errors.New("slack poster is nil")
	}
	if sleep == nil {
		sleep = time.Sleep
	}
	itemsToProcess := FeedItemsFromData(data)
	failed := 0

	for _, item := range itemsToProcess {
		// Slackメッセージ用テキストを生成
		// Descriptionは長くなる可能性があるので、Attachmentに含める
		markdownText := fmt.Sprintf("<%s|%s>", item.URL, item.Title)

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
			failed++
		}
		// レートリミット対策
		sleep(1500 * time.Millisecond)
	}
	if failed > 0 {
		return fmt.Errorf("Slackへのメッセージ送信に%d件失敗しました", failed)
	}
	return nil
}

func outputSlackDryRun(data interface{}, channel string, w io.Writer) error {
	if w == nil {
		w = io.Discard
	}
	itemsToProcess := FeedItemsFromData(data)
	if channel == "" {
		channel = "(unset)"
	}
	fmt.Fprintf(w, "slack dry-run: channel=%s posts=%d\n", channel, len(itemsToProcess))
	for i, item := range itemsToProcess {
		fmt.Fprintf(w, "%d. message title=%q url=%q\n", i+1, item.Title, item.URL)
		if item.HatenaBookmarkCount != "" {
			fmt.Fprintf(w, "   hatena_bookmark_count=%s comments=%d\n", item.HatenaBookmarkCount, len(item.HatenaBookmarkComments))
		}
	}
	return nil
}

// formatSlackAttachment は、FilteredItemからSlackのAttachmentを生成します。
func formatSlackAttachment(item FeedItem) slack.Attachment {
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
