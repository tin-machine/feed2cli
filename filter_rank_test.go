package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestKeywordFilterStage(t *testing.T) {
	items := []FeedItem{
		{Title: "Go RSS pipeline", Description: "useful"},
		{Title: "Ruby feed", Description: "skip dog"},
		{Title: "Other", Description: "nothing"},
	}
	got, err := KeywordFilterStage{
		Include:  []string{"rss", "go"},
		Exclude:  []string{"dog"},
		MinScore: 2,
	}.Apply(context.Background(), items)
	if err != nil {
		t.Fatalf("KeywordFilterStage returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Go RSS pipeline" {
		t.Fatalf("filtered items = %#v", got)
	}
	if got[0].Metadata["keyword_score"] != "2" {
		t.Fatalf("keyword score = %q", got[0].Metadata["keyword_score"])
	}
}

func TestDomainFilterStage(t *testing.T) {
	items := []FeedItem{
		{Title: "keep", URL: "https://sub.example.com/a", NormalizedURL: "https://sub.example.com/a"},
		{Title: "drop", URL: "https://evil.example.net/a", NormalizedURL: "https://evil.example.net/a"},
	}
	got, err := DomainFilterStage{Include: []string{"example.com"}}.Apply(context.Background(), items)
	if err != nil {
		t.Fatalf("DomainFilterStage returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "keep" {
		t.Fatalf("domain filtered items = %#v", got)
	}
	if got[0].Metadata["domain"] != "sub.example.com" {
		t.Fatalf("domain metadata = %q", got[0].Metadata["domain"])
	}
}

func TestTimeWindowFilterStage(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-1 * time.Hour)
	old := now.Add(-48 * time.Hour)
	items := []FeedItem{
		{Title: "recent", PublishedAt: &recent},
		{Title: "old", PublishedAt: &old},
	}
	got, err := TimeWindowFilterStage{Since: 24 * time.Hour, Now: now}.Apply(context.Background(), items)
	if err != nil {
		t.Fatalf("TimeWindowFilterStage returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "recent" {
		t.Fatalf("time filtered items = %#v", got)
	}
}

func TestHotnessFavUserAndRankStages(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-1 * time.Hour)
	older := now.Add(-24 * time.Hour)
	items := []FeedItem{
		{
			Title:               "low",
			PublishedAt:         &older,
			HatenaBookmarkCount: "1",
			HatenaBookmarkComments: []HatenaBookmarkComment{
				{User: "bob"},
			},
		},
		{
			Title:               "high",
			PublishedAt:         &recent,
			HatenaBookmarkCount: "10",
			HatenaBookmarkComments: []HatenaBookmarkComment{
				{User: "alice"},
			},
		},
	}
	got, err := RunFeedItemStages(
		context.Background(),
		items,
		HotnessScoreStage{Now: now},
		FavUserStage{Users: []string{"alice"}},
		RankStage{By: "hotness"},
	)
	if err != nil {
		t.Fatalf("RunFeedItemStages returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "high" {
		t.Fatalf("ranked items = %#v", got)
	}
	if got[0].Metadata["hotness_score"] == "" || got[0].Metadata["fav_users"] != "alice" {
		t.Fatalf("metadata missing: %#v", got[0].Metadata)
	}
}

func TestRunFilterRankFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := testRSS("go", "https://example.com/go") + testRSS("dog", "https://example.net/dog")
	code := run(
		[]string{"feed2cli", "-include-keyword", "go", "-include-domain", "example.com", "-rank", "published", "-o", "jsonl"},
		strings.NewReader(input),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, `"title":"go item"`) || strings.Contains(got, `"title":"dog item"`) {
		t.Fatalf("unexpected output:\n%s", got)
	}
}
