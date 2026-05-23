package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceLabelStage(t *testing.T) {
	got, err := SourceLabelStage{DefaultSource: "default"}.Apply(context.Background(), []FeedItem{{Title: "item"}})
	if err != nil {
		t.Fatalf("SourceLabelStage returned error: %v", err)
	}
	if got[0].Source != "default" || got[0].Metadata["source_label"] != "default" {
		t.Fatalf("source label not applied: %#v", got[0])
	}
}

func TestTagEnrichStage(t *testing.T) {
	got, err := TagEnrichStage{}.Apply(context.Background(), []FeedItem{{
		Categories: []string{"go"},
		HatenaBookmarkComments: []HatenaBookmarkComment{
			{Tags: []string{"rss", "go"}},
		},
	}})
	if err != nil {
		t.Fatalf("TagEnrichStage returned error: %v", err)
	}
	if strings.Join(got[0].Categories, ",") != "go,rss" {
		t.Fatalf("categories = %#v", got[0].Categories)
	}
}

func TestOGPEnrichStage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<!doctype html>
<html><head>
<title>fallback title</title>
<meta property="og:title" content="OG title">
<meta property="og:description" content="OG description">
<meta property="og:image" content="https://example.com/image.png">
<meta property="og:site_name" content="Example">
</head><body></body></html>`))
	}))
	defer server.Close()

	got, err := OGPEnrichStage{Client: server.Client()}.Apply(context.Background(), []FeedItem{{URL: server.URL}})
	if err != nil {
		t.Fatalf("OGPEnrichStage returned error: %v", err)
	}
	if got[0].Title != "OG title" || got[0].Description != "OG description" {
		t.Fatalf("title/description not enriched: %#v", got[0])
	}
	if got[0].Metadata["og:image"] == "" || got[0].Metadata["og:site_name"] != "Example" {
		t.Fatalf("metadata not enriched: %#v", got[0].Metadata)
	}
}

func TestContentFetchStage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<!doctype html><html><body><nav>menu</nav><article>
			<h1>Title</h1><p>First paragraph.</p><p>Second paragraph.</p>
		</article></body></html>`))
	}))
	defer server.Close()

	got, err := ContentFetchStage{Client: server.Client()}.Apply(context.Background(), []FeedItem{{URL: server.URL}})
	if err != nil {
		t.Fatalf("ContentFetchStage returned error: %v", err)
	}
	if !strings.Contains(got[0].Content, "First paragraph.") || strings.Contains(got[0].Content, "menu") {
		t.Fatalf("unexpected content: %q", got[0].Content)
	}
	if got[0].Metadata["content_text"] != got[0].Content {
		t.Fatalf("content_text metadata mismatch")
	}
}

func TestSummaryStage(t *testing.T) {
	got, err := SummaryStage{}.Apply(context.Background(), []FeedItem{{
		Description:         strings.Repeat("a", 200),
		HatenaBookmarkCount: "42",
	}})
	if err != nil {
		t.Fatalf("SummaryStage returned error: %v", err)
	}
	if got[0].Summary == "" || len([]rune(got[0].Summary)) > 160 {
		t.Fatalf("summary = %q", got[0].Summary)
	}
	if got[0].ReadingReason != "Hatena Bookmark: 42" {
		t.Fatalf("reading reason = %q", got[0].ReadingReason)
	}
}

func TestApplyEnrichments(t *testing.T) {
	got, err := applyEnrichments([]string{"source_label", "summary"}, []FeedItem{{Title: "title", Source: "feed"}})
	if err != nil {
		t.Fatalf("applyEnrichments returned error: %v", err)
	}
	if got[0].Metadata["source_label"] != "feed" || got[0].Summary == "" {
		t.Fatalf("enrichment missing: %#v", got[0])
	}
}
