package main

import (
	"sort"
	"testing"
	"github.com/mmcdole/gofeed"
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
