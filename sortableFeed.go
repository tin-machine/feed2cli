package main

import (
	"sort"

	"github.com/mmcdole/gofeed"
)

// sortableFeed は、gofeed.Feed をラップして、アイテムをソート可能にする構造体です。
type sortableFeed struct {
	gofeed.Feed
}

// Len は、フィード内のアイテムの数を返します。
func (sf sortableFeed) Len() int {
	return len(sf.Items)
}

// Swap は、二つのアイテムの位置を入れ替えます。
func (sf sortableFeed) Swap(i, j int) {
	sf.Items[i], sf.Items[j] = sf.Items[j], sf.Items[i]
}

// Less は、二つのアイテムを比較して、ソート順を決定します。
// 公開日が新しいものが先になるようにソートします。
func (sf sortableFeed) Less(i, j int) bool {
	// PublishedParsed が nil の場合は、Published を使用して比較
	// 両方 nil の場合は、等しいとみなす
	timeI := sf.Items[i].PublishedParsed
	timeJ := sf.Items[j].PublishedParsed

	if timeI == nil && timeJ == nil {
		return false // 順序は変わらない
	}
	if timeI == nil {
		return true // nil は常に大きいとみなす (降順なので)
	}
	if timeJ == nil {
		return false // nil は常に大きいとみなす (降順なので)
	}
	return timeI.After(*timeJ) // 降順
}

// Sort は sortableFeed のアイテムをソートします。
func (sf *sortableFeed) Sort() {
	sort.Sort(sf)
}
