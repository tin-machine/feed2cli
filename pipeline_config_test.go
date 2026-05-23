package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPipelineStagesFromConfig(t *testing.T) {
	stages, err := PipelineStagesFromConfig(PipelineConfig{Stages: []PipelineStageSpec{
		{Type: "normalize"},
		{Type: "keyword_filter", Include: []string{"go"}, Exclude: []string{"dog"}, MinScore: 1},
		{Type: "domain_filter", Include: []string{"example.com"}},
		{Type: "time_window", Since: "24h"},
		{Type: "rank", By: "published"},
	}})
	if err != nil {
		t.Fatalf("PipelineStagesFromConfig returned error: %v", err)
	}
	if len(stages) != 5 {
		t.Fatalf("len(stages) = %d, want 5", len(stages))
	}
	if stages[1].Name() != "keyword_filter" {
		t.Fatalf("stage[1] = %q", stages[1].Name())
	}
}

func TestPipelineStagesFromConfigRejectsUnsupportedStage(t *testing.T) {
	_, err := PipelineStagesFromConfig(PipelineConfig{Stages: []PipelineStageSpec{{Type: "unknown"}}})
	if err == nil {
		t.Fatal("PipelineStagesFromConfig returned nil error for unsupported stage")
	}
}

func TestRunWithPipelineConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "pipeline.json")
	config := `{
  "input": {"format": "feed"},
  "stages": [
    {"type": "normalize"},
    {"type": "merge"},
    {"type": "keyword_filter", "include": ["go"]},
    {"type": "domain_filter", "include": ["example.com"]}
  ],
  "output": {"type": "jsonl"}
}`
	if err := os.WriteFile(configPath, []byte(config), 0666); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	input := testRSS("go", "https://example.com/go?utm_source=x") +
		testRSS("dog", "https://example.net/dog")
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-config", configPath},
		strings.NewReader(input),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, `"schema_version":"feed2cli.feed_item.v1"`) {
		t.Fatalf("jsonl output missing schema version:\n%s", got)
	}
	if !strings.Contains(got, `"title":"go item"`) || strings.Contains(got, `"title":"dog item"`) {
		t.Fatalf("unexpected output:\n%s", got)
	}
	if !strings.Contains(got, `"normalized_url":"https://example.com/go"`) {
		t.Fatalf("normalized url missing:\n%s", got)
	}
}

func TestRunWithExternalCommandStageInPipelineConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "pipeline-plugin.json")
	config := `{
  "input": {"format": "feed"},
  "stages": [
    {
      "type": "plugin",
      "command": ` + quoteJSON(os.Args[0]) + `,
      "args": ["-test.run=TestExternalCommandStageHelper", "--", "filter-go"],
      "env": ["FEED2CLI_PLUGIN_TEST_HELPER=1"],
      "timeout": "5s"
    }
  ],
  "output": {"type": "jsonl"}
}`
	if err := os.WriteFile(configPath, []byte(config), 0666); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	input := testRSS("go", "https://example.com/go") +
		testRSS("ruby", "https://example.com/ruby")
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-config", configPath},
		strings.NewReader(input),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, `"title":"go item"`) || strings.Contains(got, `"title":"ruby item"`) {
		t.Fatalf("unexpected output:\n%s", got)
	}
}

func quoteJSON(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
