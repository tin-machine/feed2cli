package main

import (
	"testing"
	"github.com/mmcdole/gofeed"
)

func TestDiff_NormalCase(t *testing.T) {
	feedA := &gofeed.Feed{
		Items: []*gofeed.Item{
			{Title: "A", Link: "http://a/1"},
			{Title: "B", Link: "http://a/2"},
		},
	}
	feedB := &gofeed.Feed{
		Items: []*gofeed.Item{
			{Title: "B", Link: "http://a/2"},
		},
	}
	result := Diff([]*gofeed.Feed{feedA, feedB})
	if len(result) != 1 || len(result[0].Items) != 1 || result[0].Items[0].Link != "http://a/1" {
		t.Errorf("expected only A to be returned, got %v", result[0].Items)
	}
}

func TestDiff_EmptyFeeds(t *testing.T) {
	feedA := &gofeed.Feed{}
	feedB := &gofeed.Feed{}
	diff := Diff([]*gofeed.Feed{feedA, feedB})
	if len(diff) != 1 || len(diff[0].Items) != 0 {
		t.Errorf("expected empty diff, got %v", diff[0].Items)
	}
}

func TestDiff_BothFeedsNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic when feeds slice < 2")
		}
	}()
	Diff([]*gofeed.Feed{})
}

func TestDiff_FirstFeedNil(t *testing.T) {
	feedB := &gofeed.Feed{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic when first feed is nil")
		}
	}()
	Diff([]*gofeed.Feed{nil, feedB})
}

func TestDiff_SecondFeedNil(t *testing.T) {
	feedA := &gofeed.Feed{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic when second feed is nil")
		}
	}()
	Diff([]*gofeed.Feed{feedA, nil})
}

func TestDiff_AllItemsSame(t *testing.T) {
	item := &gofeed.Item{Title: "X", Link: "L"}
	feedA := &gofeed.Feed{Items: []*gofeed.Item{item}}
	feedB := &gofeed.Feed{Items: []*gofeed.Item{item}}
	diff := Diff([]*gofeed.Feed{feedA, feedB})
	if len(diff) != 1 || len(diff[0].Items) != 0 {
		t.Errorf("expected diff to be empty, got %v", diff[0].Items)
	}
}

func TestDiff_AllItemsDifferent(t *testing.T) {
	feedA := &gofeed.Feed{Items: []*gofeed.Item{{Link: "A"}}}
	feedB := &gofeed.Feed{Items: []*gofeed.Item{{Link: "B"}}}
	diff := Diff([]*gofeed.Feed{feedA, feedB})
	if len(diff) != 1 || len(diff[0].Items) != 1 || diff[0].Items[0].Link != "A" {
		t.Errorf("expected A in diff, got %v", diff[0].Items)
	}
}
