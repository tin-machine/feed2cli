package main

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedDocument struct {
	Title       string            `json:"title,omitempty"`
	Link        string            `json:"link,omitempty"`
	Description string            `json:"description,omitempty"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
	Source      string            `json:"source,omitempty"`
	Items       []FeedItem        `json:"items"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Raw         *gofeed.Feed      `json:"-"`
}

type FeedItem struct {
	ID                     string                  `json:"id,omitempty"`
	URL                    string                  `json:"url,omitempty"`
	NormalizedURL          string                  `json:"normalized_url,omitempty"`
	Title                  string                  `json:"title,omitempty"`
	Description            string                  `json:"description,omitempty"`
	Content                string                  `json:"content,omitempty"`
	PublishedAt            *time.Time              `json:"published_at,omitempty"`
	UpdatedAt              *time.Time              `json:"updated_at,omitempty"`
	Source                 string                  `json:"source,omitempty"`
	Categories             []string                `json:"categories,omitempty"`
	Metadata               map[string]string       `json:"metadata,omitempty"`
	Summary                string                  `json:"summary,omitempty"`
	ReadingReason          string                  `json:"reading_reason,omitempty"`
	HatenaBookmarkCount    string                  `json:"hatena_bookmark_count,omitempty"`
	HatenaBookmarkComments []HatenaBookmarkComment `json:"hatena_bookmark_comments,omitempty"`
	Raw                    *gofeed.Item            `json:"-"`
}

func FeedDocumentsFromFeeds(feeds []*gofeed.Feed) []FeedDocument {
	docs := make([]FeedDocument, 0, len(feeds))
	for _, feed := range feeds {
		if feed == nil {
			continue
		}
		docs = append(docs, FeedDocumentFromFeed(feed))
	}
	return docs
}

func FeedDocumentFromFeed(feed *gofeed.Feed) FeedDocument {
	doc := FeedDocument{
		Title:       feed.Title,
		Link:        feed.Link,
		Description: feed.Description,
		Source:      feed.Title,
		Items:       make([]FeedItem, 0, len(feed.Items)),
		Raw:         feed,
	}
	if feed.UpdatedParsed != nil {
		updated := *feed.UpdatedParsed
		doc.UpdatedAt = &updated
	}
	for _, item := range feed.Items {
		if item == nil {
			continue
		}
		doc.Items = append(doc.Items, FeedItemFromItem(item, doc.Source))
	}
	return doc
}

func FeedItemFromItem(item *gofeed.Item, source string) FeedItem {
	feedItem := FeedItem{
		ID:            item.GUID,
		URL:           item.Link,
		NormalizedURL: itemDedupKey(item),
		Title:         item.Title,
		Description:   item.Description,
		Content:       item.Content,
		Source:        source,
		Categories:    append([]string(nil), item.Categories...),
		Raw:           item,
	}
	if feedItem.ID == "" {
		feedItem.ID = feedItem.NormalizedURL
	}
	if item.PublishedParsed != nil {
		published := *item.PublishedParsed
		feedItem.PublishedAt = &published
	}
	if item.UpdatedParsed != nil {
		updated := *item.UpdatedParsed
		feedItem.UpdatedAt = &updated
	}
	return feedItem
}

func FeedItemsFromFilteredItems(items []*FilteredItem) []FeedItem {
	feedItems := make([]FeedItem, 0, len(items))
	for _, item := range items {
		if item == nil || item.Item == nil {
			continue
		}
		feedItem := FeedItemFromItem(item.Item, "")
		feedItem.HatenaBookmarkCount = item.HatenaBookmarkCount
		feedItem.HatenaBookmarkComments = append([]HatenaBookmarkComment(nil), item.HatenaBookmarkComments...)
		feedItems = append(feedItems, feedItem)
	}
	return feedItems
}

func FeedItemsFromData(data interface{}) []FeedItem {
	switch v := data.(type) {
	case []FeedItem:
		return v
	case []FeedDocument:
		var items []FeedItem
		for _, doc := range v {
			items = append(items, doc.Items...)
		}
		return items
	case []*FilteredItem:
		return FeedItemsFromFilteredItems(v)
	case []*gofeed.Feed:
		return FeedItemsFromDocuments(FeedDocumentsFromFeeds(v))
	case []*gofeed.Item:
		items := make([]FeedItem, 0, len(v))
		for _, item := range v {
			if item == nil {
				continue
			}
			items = append(items, FeedItemFromItem(item, ""))
		}
		return items
	default:
		return nil
	}
}

func FeedItemsFromDocuments(docs []FeedDocument) []FeedItem {
	var items []FeedItem
	for _, doc := range docs {
		items = append(items, doc.Items...)
	}
	return items
}

func (item FeedItem) PublishedTime() (time.Time, bool) {
	if item.PublishedAt != nil {
		return *item.PublishedAt, true
	}
	if item.UpdatedAt != nil {
		return *item.UpdatedAt, true
	}
	if item.Raw != nil {
		return itemPublishedTime(item.Raw)
	}
	return time.Time{}, false
}

func (item FeedItem) ToGofeedItem() *gofeed.Item {
	if item.Raw != nil {
		return item.Raw
	}
	out := &gofeed.Item{
		GUID:        item.ID,
		Title:       item.Title,
		Link:        item.URL,
		Description: item.Description,
		Content:     item.Content,
		Categories:  append([]string(nil), item.Categories...),
	}
	if item.PublishedAt != nil {
		published := *item.PublishedAt
		out.PublishedParsed = &published
		out.Published = published.Format(time.RFC3339)
	}
	if item.UpdatedAt != nil {
		updated := *item.UpdatedAt
		out.UpdatedParsed = &updated
		out.Updated = updated.Format(time.RFC3339)
	}
	return out
}

func FeedFromItems(items []FeedItem) *gofeed.Feed {
	feed := &gofeed.Feed{Items: make([]*gofeed.Item, 0, len(items))}
	for _, item := range items {
		feed.Items = append(feed.Items, item.ToGofeedItem())
	}
	return feed
}
