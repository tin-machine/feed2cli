package main

import "github.com/mmcdole/gofeed"

// Diff は、二つのフィードを受け取り、古いフィードに存在するが新しいフィードには存在しないアイテムのリストを返します。
// 引数 fs は、古いフィードが fs[0] に、新しいフィードが fs[1] に格納されていることを前提としています。
// 戻り値として、新しいフィードに含まれないアイテムだけを持つ新しいフィードを返します。
func Diff(fs []*gofeed.Feed) []*gofeed.Feed {
	if len(fs) < 2 || fs[0] == nil || fs[1] == nil {
		panic("Diff requires two non-nil feeds")
	}
	oldItems := FeedItemsFromData([]*gofeed.Feed{fs[0]})
	newItems := FeedItemsFromData([]*gofeed.Feed{fs[1]})
	diffItems := DiffFeedItems(oldItems, newItems)
	diffFeed := FeedFromItems(diffItems)
	diffFeed.FeedType = "atom"
	diffFeed.Description = "diff Atom"
	return []*gofeed.Feed{diffFeed}
}

func DiffFeedItems(oldItems, newItems []FeedItem) []FeedItem {
	newKeys := make(map[string]struct{}, len(newItems))
	for _, item := range newItems {
		key := feedItemDedupKey(item)
		if key != "" {
			newKeys[key] = struct{}{}
		}
	}

	diffItems := make([]FeedItem, 0, len(oldItems))
	for _, item := range oldItems {
		if _, exists := newKeys[feedItemDedupKey(item)]; !exists {
			diffItems = append(diffItems, item)
		}
	}
	SortFeedItems(diffItems)
	return diffItems
}
