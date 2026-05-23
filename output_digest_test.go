package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestOutputDigestMarkdownToFiltersAndSortsByWindow(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	newer := now.Add(-1 * time.Hour)
	older := now.Add(-48 * time.Hour)
	feeds := []*gofeed.Feed{{
		Items: []*gofeed.Item{
			{Title: "old", Link: "https://example.com/old", Description: "old desc", PublishedParsed: &older},
			{Title: "new", Link: "https://example.com/new", Description: "new desc", PublishedParsed: &newer},
		},
	}}

	var out bytes.Buffer
	err := OutputDigestMarkdownTo(&out, feeds, DigestOptions{
		Title:  "test digest",
		Window: 24 * time.Hour,
		Now:    now,
	})
	if err != nil {
		t.Fatalf("OutputDigestMarkdownTo returned error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "# test digest") {
		t.Fatalf("digest missing title:\n%s", got)
	}
	if !strings.Contains(got, "[new](https://example.com/new)") {
		t.Fatalf("digest missing new item:\n%s", got)
	}
	if strings.Contains(got, "https://example.com/old") {
		t.Fatalf("digest included old item outside window:\n%s", got)
	}
}
