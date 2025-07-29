package main

import (
	"github.com/mmcdole/gofeed"
)

// Merge は、引数として与えられたフィードのリストをマージし、重複を排除します。
// 戻り値として、マージされたフィードのスライスを返します。
func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	mergedFeed := &sortableFeed{gofeed.Feed{Items: []*gofeed.Item{}}} // 空のフィードを初期化

	// フィードをマージするため、アイテムを追加
	for _, v := range fs {
		for _, f := range v.Items {
			// 既存アイテムと比較
			if !itemExists(mergedFeed.Items, f.Link) {
				mergedFeed.Items = append(mergedFeed.Items, f) // 同じURLがなかったら、そのフィードを追加
			}
		}
	}

	mergedFeed.Sort()

	return []*gofeed.Feed{&mergedFeed.Feed}
}

