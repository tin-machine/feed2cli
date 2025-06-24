package main

import (
	"log"
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Merge は、引数として与えられたフィードのリストをマージし、重複を排除します。
// 戻り値として、マージされたフィードのスライスを返します。
func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()

	// マージするフィードを作成
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom"><subtitle>Example Atom</subtitle></feed>`

	mergedFeed, err := fp.Parse(strings.NewReader(feedData))
	if err != nil {
		log.Fatalf("Feed parsing failed: %v", err)
	}

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

	return []*gofeed.Feed{mergedFeed}
}

