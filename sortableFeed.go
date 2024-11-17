package main

import (
	"github.com/mmcdole/gofeed"
)

// sortableFeed は、gofeed.Feed をラップして、アイテムをソート可能にする構造体です。
type sortableFeed struct {
	gofeed.Feed
}

// Len は、フィード内のアイテムの数を返します。
func (b sortableFeed) Len() int {
	return len(b.Items)
}

// Swap は、二つのアイテムの位置を入れ替えます。
func (b sortableFeed) Swap(i, j int) {
	b.Items[i], b.Items[j] = b.Items[j], b.Items[i]
}

// Less は、二つのアイテムを比較して、ソート順を決定します。
// 公開日が新しいものが先になるようにソートします。
func (b sortableFeed) Less(i, j int) bool {
	return b.Items[i].Published > b.Items[j].Published
}
