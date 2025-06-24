package main

import (
	"testing"
	"github.com/mmcdole/gofeed"
)

func TestMerge_Normal(t *testing.T) {
	itemA := &gofeed.Item{Title: "A", Link: "a"}
	itemB := &gofeed.Item{Title: "B", Link: "b"}
	feed1 := &gofeed.Feed{Items: []*gofeed.Item{itemA, itemB}}
	feed2 := &gofeed.Feed{Items: []*gofeed.Item{itemB}}
	merged := Merge([]*gofeed.Feed{feed1, feed2})
	if len(merged) != 1 || len(merged[0].Items) != 2 {
		t.Fatalf("expected 2 unique items, got %v", merged[0].Items)
	}
	if merged[0].Items[0].Link != "a" && merged[0].Items[1].Link != "b" {
		t.Errorf("merge order might be incorrect, got %v", merged[0].Items)
	}
}

func TestMerge_EmptyInput(t *testing.T) {
	merged := Merge([]*gofeed.Feed{})
	if len(merged) != 1 || len(merged[0].Items) != 0 {
		t.Errorf("expected empty output feed, got %v", merged[0].Items)
	}
}

func TestMerge_AllDuplicate(t *testing.T) {
	itemA := &gofeed.Item{Title: "A", Link: "a"}
	feed1 := &gofeed.Feed{Items: []*gofeed.Item{itemA}}
	feed2 := &gofeed.Feed{Items: []*gofeed.Item{itemA}}
	merged := Merge([]*gofeed.Feed{feed1, feed2})
	if len(merged) != 1 || len(merged[0].Items) != 1 {
		t.Errorf("expected 1 unique item, got %v", merged[0].Items)
	}
}

func TestMerge_SingleFeed(t *testing.T) {
	itemA := &gofeed.Item{Title: "A", Link: "a"}
	feed := &gofeed.Feed{Items: []*gofeed.Item{itemA}}
	merged := Merge([]*gofeed.Feed{feed})
	if len(merged) != 1 || len(merged[0].Items) != 1 {
		t.Errorf("expected item present, got %v", merged[0].Items)
	}
}
