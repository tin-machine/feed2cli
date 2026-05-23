package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadJSONLItemsReadsStableRecords(t *testing.T) {
	var out bytes.Buffer
	items := []FeedItem{{
		Title:  "one",
		URL:    "https://example.com/one?utm_source=x",
		Source: "source",
		Metadata: map[string]string{
			"hotness_score": "10.000",
		},
	}}
	if err := OutputJSONLTo(&out, items); err != nil {
		t.Fatalf("OutputJSONLTo returned error: %v", err)
	}

	got, err := ReadJSONLItems(strings.NewReader(out.String()))
	if err != nil {
		t.Fatalf("ReadJSONLItems returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(got))
	}
	if got[0].NormalizedURL != "https://example.com/one" {
		t.Fatalf("NormalizedURL = %q", got[0].NormalizedURL)
	}
	if got[0].Metadata["hotness_score"] != "10.000" {
		t.Fatalf("metadata = %#v", got[0].Metadata)
	}
}

func TestReadJSONLItemsReadsLegacyFeedItems(t *testing.T) {
	input := `{"title":"legacy","url":"https://example.com/legacy?utm_source=x"}` + "\n"

	got, err := ReadJSONLItems(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSONLItems returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "legacy" {
		t.Fatalf("items = %#v", got)
	}
	if got[0].NormalizedURL != "https://example.com/legacy" {
		t.Fatalf("NormalizedURL = %q", got[0].NormalizedURL)
	}
}

func TestReadJSONLItemsRejectsUnsupportedSchema(t *testing.T) {
	input := `{"schema_version":"feed2cli.feed_item.v999","title":"bad"}` + "\n"

	if _, err := ReadJSONLItems(strings.NewReader(input)); err == nil {
		t.Fatal("ReadJSONLItems returned nil error for unsupported schema")
	}
}
