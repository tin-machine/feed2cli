package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExternalCommandStageFiltersJSONL(t *testing.T) {
	stage := ExternalCommandStage{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestExternalCommandStageHelper", "--", "filter-go"},
		Env:     []string{"FEED2CLI_PLUGIN_TEST_HELPER=1"},
		Timeout: 5 * time.Second,
	}
	got, err := stage.Apply(context.Background(), []FeedItem{
		{Title: "go item", URL: "https://example.com/go"},
		{Title: "ruby item", URL: "https://example.com/ruby"},
	})
	if err != nil {
		t.Fatalf("ExternalCommandStage returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "go item" {
		t.Fatalf("items = %#v", got)
	}
	if got[0].NormalizedURL != "https://example.com/go" {
		t.Fatalf("NormalizedURL = %q", got[0].NormalizedURL)
	}
}

func TestExternalCommandStageReturnsStderrOnFailure(t *testing.T) {
	stage := ExternalCommandStage{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestExternalCommandStageHelper", "--", "fail"},
		Env:     []string{"FEED2CLI_PLUGIN_TEST_HELPER=1"},
		Timeout: 5 * time.Second,
	}
	_, err := stage.Apply(context.Background(), []FeedItem{{Title: "item"}})
	if err == nil {
		t.Fatal("ExternalCommandStage returned nil error for failed command")
	}
	if !strings.Contains(err.Error(), "plugin failed intentionally") {
		t.Fatalf("error does not contain stderr: %v", err)
	}
}

func TestExternalCommandStageTimeout(t *testing.T) {
	stage := ExternalCommandStage{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestExternalCommandStageHelper", "--", "sleep"},
		Env:     []string{"FEED2CLI_PLUGIN_TEST_HELPER=1"},
		Timeout: 10 * time.Millisecond,
	}
	_, err := stage.Apply(context.Background(), []FeedItem{{Title: "item"}})
	if err == nil {
		t.Fatal("ExternalCommandStage returned nil error for timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("error = %v", err)
	}
}

func TestExternalCommandStageHelper(t *testing.T) {
	if os.Getenv("FEED2CLI_PLUGIN_TEST_HELPER") != "1" {
		return
	}
	mode := ""
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			mode = os.Args[i+1]
			break
		}
	}
	switch mode {
	case "filter-go":
		items, err := ReadJSONLItems(os.Stdin)
		if err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
			os.Exit(2)
		}
		var out []FeedItem
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), "go") {
				out = append(out, item)
			}
		}
		if err := OutputJSONLTo(os.Stdout, out); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
			os.Exit(3)
		}
		os.Exit(0)
	case "fail":
		_, _ = os.Stderr.WriteString("plugin failed intentionally")
		os.Exit(7)
	case "sleep":
		time.Sleep(time.Second)
		os.Exit(0)
	default:
		_, _ = os.Stderr.WriteString("unknown helper mode")
		os.Exit(9)
	}
}
