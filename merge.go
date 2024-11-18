package main

import (
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Merge は、引数として与えられたフィードのリストをマージし、重複を排除します。
// 戻り値として、マージされたフィードのスライスを返します。
func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()

	// マージするフィードを作成
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom">
  <subtitle>Example Atom</subtitle>
  </feed>`

	mergedFeed, _ := fp.Parse(strings.NewReader(feedData))

	// フィードをマージするため、アイテムを追加
	for _, v := range fs {
		for _, f := range v.Items {
			// 既存アイテムと比較
			if !itemExists(mergedFeed.Items, f.Link) {
				mergedFeed.Items = append(mergedFeed.Items, f) // 同じURLがなかったら、そのフィードを追加
			}
		}
	}

	sort.Sort(mergedFeed)

	output_feed := []*gofeed.Feed{mergedFeed}
	return output_feed
}

// itemExists は、アイテムのリンクが既に存在するかをチェックします。
func itemExists(items []*gofeed.Item, link string) bool {
	for _, item := range items {
		if item.Link == link {
			return true // 存在する場合はtrue
		}
	}
	return false // 存在しない場合はfalse
}
