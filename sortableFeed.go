package main

import (
	"sort"
	"time"

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
	timeI, okI := itemPublishedTime(sf.Items[i])
	timeJ, okJ := itemPublishedTime(sf.Items[j])

	if !okI && !okJ {
		return false // 順序は変わらない
	}
	if !okI {
		return false
	}
	if !okJ {
		return true
	}
	return timeI.After(timeJ) // 降順
}

// Sort は sortableFeed のアイテムをソートします。
func (sf *sortableFeed) Sort() {
	sort.Sort(sf)
}

func SortFeedItems(items []FeedItem) {
	sort.SliceStable(items, func(i, j int) bool {
		timeI, okI := items[i].PublishedTime()
		timeJ, okJ := items[j].PublishedTime()

		if !okI && !okJ {
			return false
		}
		if !okI {
			return false
		}
		if !okJ {
			return true
		}
		return timeI.After(timeJ)
	})
}

func itemPublishedTime(item *gofeed.Item) (time.Time, bool) {
	if item == nil {
		return time.Time{}, false
	}
	if item.PublishedParsed != nil {
		return *item.PublishedParsed, true
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed, true
	}

	for _, value := range []string{item.Published, item.Updated} {
		if value == "" {
			continue
		}
		for _, layout := range []string{
			time.RFC3339,
			time.RFC3339Nano,
			time.RFC1123Z,
			time.RFC1123,
			time.RFC822Z,
			time.RFC822,
			"2006-01-02 15:04:05",
			"2006-01-02",
		} {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed, true
			}
		}
	}

	return time.Time{}, false
}
