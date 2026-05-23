package main

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestNormalizeFeedURLRemovesTrackingAndCanonicalizes(t *testing.T) {
	raw := "HTTPS://M.Example.COM/posts/one/amp/?utm_source=x&b=2&fbclid=abc&a=1#frag"
	want := "https://example.com/posts/one?a=1&b=2"
	if got := normalizeFeedURL(raw); got != want {
		t.Fatalf("normalizeFeedURL() = %q, want %q", got, want)
	}
}

func TestMergeUsesNormalizedURLForDedupe(t *testing.T) {
	feed1 := &gofeed.Feed{Items: []*gofeed.Item{{Title: "raw", Link: "https://example.com/post?utm_source=x"}}}
	feed2 := &gofeed.Feed{Items: []*gofeed.Item{{Title: "canonical", Link: "https://example.com/post"}}}

	merged := Merge([]*gofeed.Feed{feed1, feed2})
	if len(merged) != 1 || len(merged[0].Items) != 1 {
		t.Fatalf("merged item count = %d, want 1", len(merged[0].Items))
	}
	if merged[0].Items[0].Link != "https://example.com/post?utm_source=x" {
		t.Fatalf("merge changed visible URL to %q", merged[0].Items[0].Link)
	}
}

func TestDiffUsesNormalizedURLForDedupe(t *testing.T) {
	oldFeed := &gofeed.Feed{Items: []*gofeed.Item{{Title: "old", Link: "https://example.com/post?utm_campaign=x"}}}
	newFeed := &gofeed.Feed{Items: []*gofeed.Item{{Title: "new", Link: "https://example.com/post"}}}

	diff := Diff([]*gofeed.Feed{oldFeed, newFeed})
	if len(diff) != 1 || len(diff[0].Items) != 0 {
		t.Fatalf("diff item count = %d, want 0", len(diff[0].Items))
	}
}
