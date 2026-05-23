package main

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestFeedDocumentFromFeedPreservesRawAndNormalizedView(t *testing.T) {
	published := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	feed := &gofeed.Feed{
		Title: "source",
		Link:  "https://example.com/feed",
		Items: []*gofeed.Item{{
			GUID:            "guid-1",
			Title:           "item",
			Link:            "https://example.com/post?utm_source=x",
			Description:     "description",
			PublishedParsed: &published,
			Categories:      []string{"go"},
		}},
	}

	doc := FeedDocumentFromFeed(feed)
	if doc.Raw != feed {
		t.Fatal("FeedDocument did not preserve raw feed")
	}
	if len(doc.Items) != 1 {
		t.Fatalf("doc item count = %d, want 1", len(doc.Items))
	}
	item := doc.Items[0]
	if item.Raw != feed.Items[0] {
		t.Fatal("FeedItem did not preserve raw item")
	}
	if item.NormalizedURL != "https://example.com/post" {
		t.Fatalf("NormalizedURL = %q", item.NormalizedURL)
	}
	if item.ID != "guid-1" {
		t.Fatalf("ID = %q", item.ID)
	}
	if item.Source != "source" {
		t.Fatalf("Source = %q", item.Source)
	}
}

func TestFeedItemsFromFilteredItemsPreservesEnrichment(t *testing.T) {
	items := FeedItemsFromFilteredItems([]*FilteredItem{{
		Item:                &gofeed.Item{Title: "item", Link: "https://example.com/item"},
		HatenaBookmarkCount: "12",
		HatenaBookmarkComments: []HatenaBookmarkComment{
			{User: "alice", Comment: "nice"},
		},
	}})
	if len(items) != 1 {
		t.Fatalf("len(items) = %d", len(items))
	}
	if items[0].HatenaBookmarkCount != "12" || len(items[0].HatenaBookmarkComments) != 1 {
		t.Fatalf("enrichment was not preserved: %#v", items[0])
	}
}

func TestMergeFeedItemsUsesInternalModel(t *testing.T) {
	items := []FeedItem{
		{Title: "a", URL: "https://example.com/post?utm_source=x", NormalizedURL: "https://example.com/post"},
		{Title: "b", URL: "https://example.com/post", NormalizedURL: "https://example.com/post"},
	}
	merged := MergeFeedItems(items)
	if len(merged) != 1 {
		t.Fatalf("len(merged) = %d, want 1", len(merged))
	}
	if merged[0].Title != "a" {
		t.Fatalf("merged first title = %q", merged[0].Title)
	}
}

func TestDiffFeedItemsUsesInternalModel(t *testing.T) {
	oldItems := []FeedItem{{Title: "old", URL: "https://example.com/post?utm_source=x", NormalizedURL: "https://example.com/post"}}
	newItems := []FeedItem{{Title: "new", URL: "https://example.com/post", NormalizedURL: "https://example.com/post"}}
	if diff := DiffFeedItems(oldItems, newItems); len(diff) != 0 {
		t.Fatalf("diff = %#v, want empty", diff)
	}
}
