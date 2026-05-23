package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestStoreFeedWithOptionsStoresAndMergesJSON(t *testing.T) {
	dir := t.TempDir()
	oldTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)

	fetch := func(string) (*gofeed.Feed, error) {
		return &gofeed.Feed{
			Link: "https://example.com/feed",
			Items: []*gofeed.Item{
				{Title: "new", Link: "https://example.com/new", PublishedParsed: &newTime},
			},
		}, nil
	}

	path, err := StoreFeedWithOptions("https://example.com/rss", StoreFeedOptions{Prefix: dir, Fetch: fetch})
	if err != nil {
		t.Fatalf("StoreFeedWithOptions returned error: %v", err)
	}

	storedOld := gofeed.Feed{
		Link: "https://example.com/feed",
		Items: []*gofeed.Item{
			{Title: "old", Link: "https://example.com/old", PublishedParsed: &oldTime},
			{Title: "duplicate", Link: "https://example.com/new?utm_source=x", PublishedParsed: &oldTime},
		},
	}
	data, err := json.Marshal(storedOld)
	if err != nil {
		t.Fatalf("failed to marshal old feed: %v", err)
	}
	if err := os.WriteFile(path, data, 0666); err != nil {
		t.Fatalf("failed to write old feed: %v", err)
	}

	path, err = StoreFeedWithOptions("https://example.com/rss", StoreFeedOptions{Prefix: dir, Fetch: fetch})
	if err != nil {
		t.Fatalf("second StoreFeedWithOptions returned error: %v", err)
	}
	gotData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read stored feed: %v", err)
	}
	var got gofeed.Feed
	if err := json.Unmarshal(gotData, &got); err != nil {
		t.Fatalf("stored feed is not JSON: %v\n%s", err, gotData)
	}
	if len(got.Items) != 2 {
		t.Fatalf("stored item count = %d, want 2: %#v", len(got.Items), got.Items)
	}
	if got.Items[0].Link != "https://example.com/new" {
		t.Fatalf("first item link = %q, want newest item first", got.Items[0].Link)
	}
}

func TestStoreFeedWithOptionsReturnsFetchError(t *testing.T) {
	_, err := StoreFeedWithOptions("https://example.com/rss", StoreFeedOptions{
		Prefix: t.TempDir(),
		Fetch: func(string) (*gofeed.Feed, error) {
			return nil, errors.New("network down")
		},
	})
	if err == nil {
		t.Fatal("StoreFeedWithOptions returned nil error")
	}
}

func TestStoreFeedPath(t *testing.T) {
	file, dir := storeFeedPath("/tmp/feed-store", &gofeed.Feed{Link: "https://example.com/news/rss"})
	if file != "/tmp/feed-store/example.com/news/rss" {
		t.Fatalf("file = %q", file)
	}
	if dir != "/tmp/feed-store/example.com/news/" {
		t.Fatalf("dir = %q", dir)
	}
}

func TestLoadStoredFeedReturnsBrokenJSONError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "feed.json")
	if err := os.WriteFile(file, []byte("{broken"), 0666); err != nil {
		t.Fatalf("failed to write broken JSON: %v", err)
	}
	if _, err := loadStoredFeed(file); err == nil {
		t.Fatal("loadStoredFeed returned nil error for broken JSON")
	}
}
