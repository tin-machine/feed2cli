package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestOutputJSONLTo(t *testing.T) {
	feeds := []*gofeed.Feed{{
		Title: "source",
		Items: []*gofeed.Item{
			{Title: "one", Link: "https://example.com/one?utm_source=x"},
			{Title: "two", Link: "https://example.com/two"},
		},
	}}

	var out bytes.Buffer
	if err := OutputJSONLTo(&out, feeds); err != nil {
		t.Fatalf("OutputJSONLTo returned error: %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	var lines []FeedItemJSONLRecord
	for scanner.Scan() {
		var item FeedItemJSONLRecord
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			t.Fatalf("line is not FeedItem JSONL record: %v", err)
		}
		lines = append(lines, item)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2\n%s", len(lines), out.String())
	}
	if lines[0].SchemaVersion != feedItemJSONLSchemaVersion {
		t.Fatalf("SchemaVersion = %q", lines[0].SchemaVersion)
	}
	if lines[0].NormalizedURL != "https://example.com/one" {
		t.Fatalf("NormalizedURL = %q", lines[0].NormalizedURL)
	}
}

func TestOutputJSONLToUnsupportedType(t *testing.T) {
	var out bytes.Buffer
	if err := OutputJSONLTo(&out, "unsupported"); err == nil {
		t.Fatal("OutputJSONLTo returned nil error for unsupported type")
	}
}
