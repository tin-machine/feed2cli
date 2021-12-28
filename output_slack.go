package main

/*
Slackに出力する

todo
* https://zenn.dev/kou_pg_0131/articles/slack-go-usage を参考にSlack通知のデザインを変えていく
*/

import (
	"log"
	"time"
  "os"
  "fmt"
  "regexp"
  "net/url"

	"github.com/gorilla/feeds"
  "github.com/mmcdole/gofeed"
  "github.com/slack-go/slack"
  "github.com/k0kubun/pp"
)


func toSlack(feed []*gofeed.Feed) {

  // アクセストークンを使用してクライアントを生成する。
  // 環境変数 XOXB に xoxb- から始まるトークンを設定しておくこと
  c := slack.New(os.Getenv("XOXB"))

  for _, f := range feed {

	  for _, v := range f.Items {
      /*
      ブックマークコメントURL Items.Extensions["hatena"]["bookmarkCommentListPageUrl"][0].Value
      関連サイト Items.Extensions["hatena"]["bookmarkSiteEntriesListUrl"][0].Value
      画像 Items.Extensions["hatena"]["imageurl"][0].Value
      */

      /*
      タグURLの処理
      */
      re := regexp.MustCompile(`.*?q=(.*)`)
      li := v.Extensions["taxo"]["topics"][0].Children["Bag"][0].Children["li"]
      tag_url := ""
      for _, li_v := range li {
        res := re.FindAllStringSubmatch(li_v.Attrs["resource"], -1)
        pp.Print(res)
        fmt.Printf("%#v\n", li_v.Attrs["resource"])
        str2, _ := url.QueryUnescape(res[0][1])
        tag_url = tag_url + "<" + li_v.Attrs["resource"] + "|" + str2 + "> "
      }

      markdownText := "<" + v.Link + "|" + v.Title + ">\n" + v.Description

      /*
      Attachmentsの方式でやってみる
      https://qiita.com/rshibasa/items/c0fc1dfc2a920852ebc3
      */
      attachment := slack.Attachment{
        ThumbURL: v.Extensions["hatena"]["imageurl"][0].Value,
        Text:    "<" + v.Extensions["hatena"]["bookmarkCommentListPageUrl"][0].Value + "|コメント> <" + v.Extensions["hatena"]["bookmarkSiteEntriesListUrl"][0].Value + "|関連>\n" + tag_url,
      }
      attach := slack.MsgOptionAttachments(attachment)

      channelID, timestamp, err := c.PostMessage("@kaoru-inoue", slack.MsgOptionText(markdownText, false), attach,  slack.MsgOptionAsUser(true))
      fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)

	    if err != nil {
	    	log.Println(err)
        // ここで、invalid_blocksと出た場合のRSSエントリのオブジェクトもダンプできるようにする。
        fmt.Printf("%#v\n", v)
	    }
      time.Sleep(3 * time.Second)
	  }
  }
}

/*
slackに出力
*/
func OutputSlack(feed []*gofeed.Feed) {

  for _, f := range feed {
	  c1 := f
	  now := time.Now()
	  output_feed := &feeds.Feed{
	  	Title:       c1.Title,
	  	Link:        &feeds.Link{Href: c1.Link},
	  	Description: c1.Description,
	  	Created:     now,
	  }
	  for _, v := range c1.Items {
	  	item := &feeds.Item{
	  		Title:       v.Title,
	  		Link:        &feeds.Link{Href: v.Link},
	  		Description: v.Description,
	  		Created:     now,
	  	}
	  	output_feed.Add(item)
	  }
	  rss, err := output_feed.ToRss()
	  if err != nil {
	  	log.Fatal(err)
	  }
    fmt.Print(rss)
    toSlack(feed)
  }

}
