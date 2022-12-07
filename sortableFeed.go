package main

import (
	"github.com/mmcdole/gofeed"
)

type sortableFeed struct {
	gofeed.Feed
}

func (b sortableFeed) Len() int {
	return len(b.Items)
}

func (b sortableFeed) Swap(i, j int) {
	b.Items[i], b.Items[j] = b.Items[j], b.Items[i]
}

func (b sortableFeed) Less(i, j int) bool {
	return b.Items[i].Published > b.Items[j].Published
}
