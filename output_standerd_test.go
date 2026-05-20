package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestOutputStandardToFeeds(t *testing.T) {
	feed := &gofeed.Feed{
		Items: []*gofeed.Item{
			{
				Title:       "entry one",
				Link:        "https://example.com/entry-one",
				Description: "entry description",
			},
		},
	}
	var out bytes.Buffer

	if err := OutputStandardTo(&out, []*gofeed.Feed{feed}, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("OutputStandardTo returned error: %v", err)
	}

	parsed, err := gofeed.NewParser().ParseString(out.String())
	if err != nil {
		t.Fatalf("generated RSS did not parse: %v\n%s", err, out.String())
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("len(parsed.Items) = %d, want 1", len(parsed.Items))
	}
	if parsed.Items[0].Title != "entry one" {
		t.Fatalf("item title = %q, want entry one", parsed.Items[0].Title)
	}
	if parsed.Items[0].Link != "https://example.com/entry-one" {
		t.Fatalf("item link = %q", parsed.Items[0].Link)
	}
}

func TestOutputStandardToFilteredItems(t *testing.T) {
	item := &FilteredItem{
		Item: &gofeed.Item{
			Title:       "filtered entry",
			Link:        "https://example.com/filtered",
			Description: "filtered description",
		},
		HatenaBookmarkCount: "12",
		HatenaBookmarkComments: []HatenaBookmarkComment{
			{User: "alice", Timestamp: "2026/05/20 12:00", Comment: "useful"},
			{User: "bob", Timestamp: "2026/05/20 12:01"},
		},
	}
	var out bytes.Buffer

	if err := OutputStandardTo(&out, []*FilteredItem{item}, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("OutputStandardTo returned error: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Hatena Bookmark:",
		"12",
		"alice",
		"useful",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated RSS does not contain %q:\n%s", want, got)
		}
	}
}

func TestOutputStandardToUnsupportedType(t *testing.T) {
	var out bytes.Buffer

	if err := OutputStandardTo(&out, "unsupported", time.Now()); err == nil {
		t.Fatal("OutputStandardTo returned nil error for unsupported type")
	}
	if out.Len() != 0 {
		t.Fatalf("output length = %d, want 0", out.Len())
	}
}
