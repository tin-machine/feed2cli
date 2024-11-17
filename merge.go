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
		// フィードの中のアイテム
		for _, f := range v.Items {
			addFlag := true
			// 既存アイテムと比較
			for j, _ := range mergedFeed.Items {
				if mergedFeed.Items[j].Link == f.Link {
					// 同じURLのフィードは追加しない、次のフィードのチェックを行う
					addFlag = false
					break
				}
			}
			// 同じURLがなかった( trueのままだった )ら、そのフィードを追加
			if addFlag {
				mergedFeed.Items = append(mergedFeed.Items, f)
			}
		}
	}
	sort.Sort(mergedFeed)

	output_feed := []*gofeed.Feed{mergedFeed}
	return output_feed
}
