package main

/*
Slackに出力する

todo
* https://zenn.dev/kou_pg_0131/articles/slack-go-usage を参考にSlack通知のデザインを変えていく
*/

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

// toSlack は、フィードを受け取り、各アイテムをSlackに送信します。
func toSlack(feed []*gofeed.Feed) {
	token := os.Getenv("XOXB")
	if token == "" {
		log.Fatal("環境変数XOXBにアクセストークンを設定してください。")
	}
	c := slack.New(token)

	for _, f := range feed {
		for _, v := range f.Items {
			tagURL := processTags(v) // タグURLの処理

			// Slackメッセージ用テキストを生成
			markdownText := fmt.Sprintf("<%s|%s>\n%s", v.Link, v.Title, v.Description)

			// Attachmentを生成
			attachment := slack.Attachment{
				ThumbURL: v.Extensions["hatena"]["imageurl"][0].Value,
				Text:     formatAttachmentText(v, tagURL),
			}

			// Slackへメッセージを送信
			attach := slack.MsgOptionAttachments(attachment)
			sendSlackMessage(c, markdownText, attach)
		}
	}
}

// processTags は、フィードアイテム内のtagsを処理します。
func processTags(v *gofeed.Item) string {
	re := regexp.MustCompile(`.*?q=(.*)`)
	li := v.Extensions["taxo"]["topics"][0].Children["Bag"][0].Children["li"]
	tagURL := ""

	for _, liV := range li {
		res := re.FindAllStringSubmatch(liV.Attrs["resource"], -1)
		if len(res) > 0 {
			if str2, err := url.QueryUnescape(res[0][1]); err == nil {
				tagURL += fmt.Sprintf("<%s|%s> ", liV.Attrs["resource"], str2)
			} else {
				log.Printf("タグURLのデコードに失敗しました: %v", err)
			}
		}
	}
	return tagURL
}

// formatAttachmentText は、Slackに送信するための添付テキストを整形します。
func formatAttachmentText(v *gofeed.Item, tagURL string) string {
	bookmarkCommentURL := v.Extensions["hatena"]["bookmarkCommentListPageUrl"][0].Value
	bookmarkSiteURL := v.Extensions["hatena"]["bookmarkSiteEntriesListUrl"][0].Value

	if bookmarkCommentURL == "" || bookmarkSiteURL == "" {
		log.Println("コメントURLまたはサイトURLが空です。")
	}

	return fmt.Sprintf("<%s|コメント> <%s|関連>\n%s",
		bookmarkCommentURL,
		tagURL,
		bookmarkSiteURL)
}

// sendSlackMessage は、Slackにメッセージを送信します。
func sendSlackMessage(c *slack.Client, markdownText string, attach slack.MsgOption) {
	channel := os.Getenv("SLACK_CHANNEL")
	if channel == "" {
		log.Fatal("環境変数SLACK_CHANNELに通知先のチャンネル名またはユーザーIDを設定してください。")
	}
	channelID, timestamp, err := c.PostMessage(channel, slack.MsgOptionText(markdownText, false), attach, slack.MsgOptionAsUser(true))
	if err != nil {
		log.Printf("Slackへのメッセージ送信でエラーが発生しました: %v", err)
	} else {
		fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
	}
}

// OutputSlack は、フィードを受け取り、標準出力およびSlackに出力します。
func OutputSlack(feed []*gofeed.Feed) {
	for _, f := range feed {
		now := time.Now()
		outputFeed := &feeds.Feed{
			Title:       f.Title,
			Link:        &feeds.Link{Href: f.Link},
			Description: f.Description,
			Created:     now,
		}
		for _, v := range f.Items {
			item := &feeds.Item{
				Title:       v.Title,
				Link:        &feeds.Link{Href: v.Link},
				Description: v.Description,
				Created:     now,
			}
			outputFeed.Add(item)
		}

		rss, err := outputFeed.ToRss()
		if err != nil {
			log.Fatalf("RSS生成に失敗しました: %v", err)
		}
		fmt.Print(rss)

		// Slackへも出力
		toSlack(feed) // ここでtoSlackを呼ぶ
	}
}
