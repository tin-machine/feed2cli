package main

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestConvertToFeeds(t *testing.T) {
	feed := &gofeed.Feed{Items: []*gofeed.Item{{Title: "feed item", Link: "feed"}}}
	filtered := []*FilteredItem{{Item: &gofeed.Item{Title: "filtered item", Link: "filtered"}}}
	items := []*gofeed.Item{{Title: "item", Link: "item"}}

	tests := []struct {
		name     string
		input    interface{}
		wantLen  int
		wantLink string
		wantNil  bool
	}{
		{name: "feeds", input: []*gofeed.Feed{feed}, wantLen: 1, wantLink: "feed"},
		{name: "filtered items", input: filtered, wantLen: 1, wantLink: "filtered"},
		{name: "items", input: items, wantLen: 1, wantLink: "item"},
		{name: "unsupported", input: "unsupported", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToFeeds(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("convertToFeeds = %v, want nil", got)
				}
				return
			}
			if len(got) != 1 || len(got[0].Items) != tt.wantLen {
				t.Fatalf("convertToFeeds returned %#v", got)
			}
			if got[0].Items[0].Link != tt.wantLink {
				t.Fatalf("link = %q, want %q", got[0].Items[0].Link, tt.wantLink)
			}
		})
	}
}

func TestConvertToFilteredItems(t *testing.T) {
	filtered := []*FilteredItem{{Item: &gofeed.Item{Title: "already filtered"}}}

	if got := convertToFilteredItems(filtered); len(got) != 1 || got[0] != filtered[0] {
		t.Fatalf("convertToFilteredItems did not preserve filtered items: %#v", got)
	}

	feeds := []*gofeed.Feed{{Items: []*gofeed.Item{{Title: "feed item"}}}}
	got := convertToFilteredItems(feeds)
	if len(got) != 1 || got[0].Title != "feed item" {
		t.Fatalf("convertToFilteredItems feeds = %#v", got)
	}

	items := []*gofeed.Item{{Title: "item"}}
	got = convertToFilteredItems(items)
	if len(got) != 1 || got[0].Title != "item" {
		t.Fatalf("convertToFilteredItems items = %#v", got)
	}

	got = convertToFilteredItems("unsupported")
	if len(got) != 0 {
		t.Fatalf("len(convertToFilteredItems unsupported) = %d, want 0", len(got))
	}
}

func TestItemExists(t *testing.T) {
	items := []*gofeed.Item{{Link: "a"}, {Link: "b"}}

	if !itemExists(items, "a") {
		t.Fatal("itemExists did not find existing link")
	}
	if itemExists(items, "missing") {
		t.Fatal("itemExists found missing link")
	}
}
