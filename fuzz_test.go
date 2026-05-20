package main

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func FuzzSplitFeed(f *testing.F) {
	for _, seed := range []string{
		"",
		"<rss></rss>",
		"<feed></feed>",
		"not xml",
		"<rss><channel></channel></rss><feed></feed>",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = splitFeed([]byte(input), true)
	})
}

func FuzzMergeUniqueLinks(f *testing.F) {
	f.Add("a", "b", "a")
	f.Add("", "b", "")
	f.Add("same", "same", "same")

	f.Fuzz(func(t *testing.T, a, b, c string) {
		merged := Merge([]*gofeed.Feed{
			{Items: []*gofeed.Item{{Link: a}, {Link: b}}},
			{Items: []*gofeed.Item{{Link: c}}},
		})
		seen := map[string]bool{}
		for _, item := range merged[0].Items {
			if seen[item.Link] {
				t.Fatalf("duplicate link %q in merged feed", item.Link)
			}
			seen[item.Link] = true
		}
	})
}

func FuzzDiffOnlyOldLinks(f *testing.F) {
	f.Add("old-only", "shared")
	f.Add("", "shared")

	f.Fuzz(func(t *testing.T, oldOnly, shared string) {
		diff := Diff([]*gofeed.Feed{
			{Items: []*gofeed.Item{{Link: oldOnly}, {Link: shared}}},
			{Items: []*gofeed.Item{{Link: shared}}},
		})
		for _, item := range diff[0].Items {
			if item.Link == shared {
				t.Fatalf("diff returned shared link %q", shared)
			}
		}
	})
}
