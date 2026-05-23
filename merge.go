package main

import (
	"github.com/mmcdole/gofeed"
)

// Merge は、引数として与えられたフィードのリストをマージし、重複を排除します。
// 戻り値として、マージされたフィードのスライスを返します。
func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	return []*gofeed.Feed{FeedFromItems(MergeFeedItems(FeedItemsFromData(fs)))}
}

func MergeFeedItems(items []FeedItem) []FeedItem {
	seen := make(map[string]struct{}, len(items))
	merged := make([]FeedItem, 0, len(items))
	for _, item := range items {
		key := item.NormalizedURL
		if key == "" {
			key = normalizeFeedURL(item.URL)
		}
		if key == "" {
			key = item.URL
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, item)
	}
	SortFeedItems(merged)
	return merged
}
