package main

import (
	"github.com/mmcdole/gofeed"
	"sort"
	"testing"
	"time"
)

func TestSortableFeed_Empty(t *testing.T) {
	sf := sortableFeed{gofeed.Feed{Items: []*gofeed.Item{}}}
	if sf.Len() != 0 {
		t.Errorf("expected 0, got %d", sf.Len())
	}
}

func TestSortableFeed_SortOrder(t *testing.T) {
	sf := sortableFeed{gofeed.Feed{
		Items: []*gofeed.Item{
			{Title: "C", Published: "2023-06-01T00:00:00Z"},
			{Title: "A", Published: "2023-08-01T00:00:00Z"},
			{Title: "B", Published: "2023-07-01T00:00:00Z"},
		},
	}}
	sort.Sort(sf)
	if sf.Items[0].Title != "A" || sf.Items[1].Title != "B" || sf.Items[2].Title != "C" {
		t.Errorf("unexpected order: %v, %v, %v", sf.Items[0].Title, sf.Items[1].Title, sf.Items[2].Title)
	}
}

func TestSortableFeed_SingleItem(t *testing.T) {
	sf := sortableFeed{gofeed.Feed{Items: []*gofeed.Item{{Title: "Z", Published: "2023-06-01T00:00:00Z"}}}}
	sort.Sort(sf)
	if sf.Len() != 1 || sf.Items[0].Title != "Z" {
		t.Errorf("unexpected result for single item: %v", sf.Items)
	}
}

func TestSortableFeed_TimeSources(t *testing.T) {
	publishedParsed := time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	updatedParsed := time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC)
	sf := sortableFeed{gofeed.Feed{
		Items: []*gofeed.Item{
			{Title: "invalid", Published: "not a time"},
			{Title: "rfc1123", Published: "Mon, 02 Jan 2023 15:04:05 GMT"},
			{Title: "updated parsed", UpdatedParsed: &updatedParsed},
			{Title: "published string", Published: "2023-06-01T00:00:00Z"},
			{Title: "published parsed wins", Published: "2020-01-01T00:00:00Z", PublishedParsed: &publishedParsed},
			nil,
		},
	}}

	sort.Sort(sf)

	wantOrder := []string{
		"published parsed wins",
		"updated parsed",
		"published string",
		"rfc1123",
		"invalid",
	}
	for i, want := range wantOrder {
		if sf.Items[i].Title != want {
			t.Fatalf("item %d = %q, want %q", i, sf.Items[i].Title, want)
		}
	}
	if sf.Items[len(sf.Items)-1] != nil {
		t.Fatalf("last item = %#v, want nil", sf.Items[len(sf.Items)-1])
	}
}
