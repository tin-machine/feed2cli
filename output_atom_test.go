package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestOutputAtomTo(t *testing.T) {
	published := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	feeds := []*gofeed.Feed{{
		Items: []*gofeed.Item{{
			GUID:            "item-id",
			Title:           "atom item",
			Link:            "https://example.com/atom",
			Description:     "description",
			PublishedParsed: &published,
		}},
	}}

	var out bytes.Buffer
	if err := OutputAtomTo(&out, feeds, published); err != nil {
		t.Fatalf("OutputAtomTo returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "<feed") || !strings.Contains(got, "atom item") {
		t.Fatalf("unexpected Atom output:\n%s", got)
	}
	parsed, err := gofeed.NewParser().ParseString(got)
	if err != nil {
		t.Fatalf("Atom output did not parse: %v\n%s", err, got)
	}
	if len(parsed.Items) != 1 || parsed.Items[0].Link != "https://example.com/atom" {
		t.Fatalf("parsed Atom items = %#v", parsed.Items)
	}
}

func TestOutputAtomToUnsupportedType(t *testing.T) {
	var out bytes.Buffer
	if err := OutputAtomTo(&out, "unsupported", time.Now()); err == nil {
		t.Fatal("OutputAtomTo returned nil error for unsupported type")
	}
}
