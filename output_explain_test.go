package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExplainItemsFilterRank(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-1 * time.Hour)
	items := []FeedItem{
		{Title: "go item", URL: "https://example.com/go", PublishedAt: &recent},
		{Title: "dog item", URL: "https://example.net/dog", PublishedAt: &recent},
	}
	cfg := cliConfig{
		operation:       "jsonl",
		includeKeyword:  []string{"go"},
		excludeKeyword:  []string{"dog"},
		includeDomain:   []string{"example.com"},
		minKeywordScore: 1,
		since:           24 * time.Hour,
	}

	got := ExplainItems(items, cfg, "jsonl", now)
	if len(got) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(got))
	}
	if !got[0].Kept {
		t.Fatalf("first record should be kept: %#v", got[0])
	}
	if got[1].Kept {
		t.Fatalf("second record should be dropped: %#v", got[1])
	}
	if !strings.Contains(strings.Join(got[1].Reasons, "\n"), "keyword_filter: dropped by exclude=dog") {
		t.Fatalf("drop reason missing: %#v", got[1].Reasons)
	}
}

func TestOutputExplainJSONLTo(t *testing.T) {
	var out bytes.Buffer
	cfg := cliConfig{operation: "digest", includeKeyword: []string{"go"}}
	if err := OutputExplainJSONLTo(&out, []FeedItem{{Title: "go", URL: "https://example.com/go"}}, cfg, "feed2cli"); err != nil {
		t.Fatalf("OutputExplainJSONLTo returned error: %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	if !scanner.Scan() {
		t.Fatalf("no explain output:\n%s", out.String())
	}
	var record ExplainRecord
	if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
		t.Fatalf("line is not explain JSON: %v", err)
	}
	if record.SchemaVersion != explainJSONLSchemaVersion {
		t.Fatalf("SchemaVersion = %q", record.SchemaVersion)
	}
	if record.Output != "digest" || !record.Kept {
		t.Fatalf("record = %#v", record)
	}
}

func TestExplainItemsPipelineEnrichAndPluginDryRun(t *testing.T) {
	items := []FeedItem{
		{Title: "go item", URL: "https://example.com/go?utm_source=x"},
		{Title: "dog item", URL: "https://example.com/dog?utm_source=x"},
		{Title: "go duplicate", URL: "https://example.com/go"},
	}
	cfg := cliConfig{
		pipelineStages: []FeedItemStage{
			NormalizeStage{},
			MergeStage{},
			KeywordFilterStage{Include: []string{"go"}},
			ExternalCommandStage{Command: "/bin/feed2cli-plugin", Args: []string{"--noop"}, Timeout: time.Second},
		},
		enrichTypes: []string{"source_label", "ogp", "summary"},
	}

	got := ExplainItems(items, cfg, "jsonl", time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC))
	if len(got) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(got))
	}
	if !got[0].Kept {
		t.Fatalf("first record should be kept: %#v", got[0])
	}
	if got[1].Kept {
		t.Fatalf("dog record should be dropped: %#v", got[1])
	}
	if got[2].Kept {
		t.Fatalf("duplicate record should be dropped: %#v", got[2])
	}

	allReasons := strings.Join(append(append([]string{}, got[0].Reasons...), got[2].Reasons...), "\n")
	for _, want := range []string{
		"merge: dropped duplicate",
		"plugin: dry-run command=/bin/feed2cli-plugin",
		"ogp: would fetch article URL",
		"summary: local extractive summary",
	} {
		if !strings.Contains(allReasons, want) {
			t.Fatalf("missing reason %q:\n%s", want, allReasons)
		}
	}
	if len(got[0].Stages) == 0 {
		t.Fatalf("stage results missing: %#v", got[0])
	}
}
