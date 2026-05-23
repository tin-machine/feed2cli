package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunJSONWritesToInjectedStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-o", "json"},
		strings.NewReader(testRSS("json", "https://example.com/json")),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"title": "json item"`) {
		t.Fatalf("stdout does not contain JSON item:\n%s", stdout.String())
	}
}

func TestRunTerminalCreatesSymlinks(t *testing.T) {
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	code := run([]string{"feed2cli", "-s"}, strings.NewReader(""), &stdout, &stderr, true)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	for _, name := range []string{"mergeRss", "diffRss", "slackRss", "hatenaRss"} {
		if !fileExists(name) {
			t.Fatalf("expected symlink %s to exist", name)
		}
	}
}

func TestRunFetchesURLWhenTerminal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRSS("url", "https://example.com/url")))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := run([]string{"feed2cli", "-url", server.URL, "-o", "jsonl"}, strings.NewReader(""), &stdout, &stderr, true)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"title":"url item"`) {
		t.Fatalf("stdout does not contain URL feed item:\n%s", stdout.String())
	}
}

func TestRunCombinesStdinAndURLFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRSS("url", "https://example.com/url")))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-url", server.URL, "-o", "jsonl"},
		strings.NewReader(testRSS("stdin", "https://example.com/stdin")),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, `"title":"stdin item"`) || !strings.Contains(got, `"title":"url item"`) {
		t.Fatalf("stdout does not contain both stdin and URL items:\n%s", got)
	}
}

func TestRunURLLint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRSS("url", "https://example.com/url")))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := run([]string{"feed2cli", "-url", server.URL, "-o", "lint"}, strings.NewReader(""), &stdout, &stderr, true)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "feeds: total=1 valid=1 invalid=0") {
		t.Fatalf("unexpected lint output:\n%s", stdout.String())
	}
}

func TestRunSlackDryRunDoesNotRequireEnv(t *testing.T) {
	t.Setenv("FEED2CLI_SLACK_TOKEN", "")
	t.Setenv("FEED2CLI_SLACK_CHANNEL", "")
	t.Setenv("XOXB", "")
	t.Setenv("SLACK_CHANNEL", "")

	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-o", "slack", "-slack-dry-run", "-slack-channel", "C123"},
		strings.NewReader(testRSS("slack", "https://example.com/slack")),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "slack dry-run: channel=C123 posts=1") {
		t.Fatalf("unexpected dry-run output:\n%s", stdout.String())
	}
}

func TestRunAcceptsJSONLInput(t *testing.T) {
	var jsonl bytes.Buffer
	if err := OutputJSONLTo(&jsonl, []FeedItem{{
		Title: "jsonl",
		URL:   "https://example.com/jsonl?utm_source=x",
	}}); err != nil {
		t.Fatalf("OutputJSONLTo returned error: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-input", "jsonl", "-include-domain", "example.com", "-o", "atom"},
		strings.NewReader(jsonl.String()),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "<title>jsonl</title>") {
		t.Fatalf("unexpected atom output:\n%s", got)
	}
}

func TestRunExplainOutputsDroppedItems(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-explain", "-include-keyword", "go", "-exclude-keyword", "dog", "-o", "jsonl"},
		strings.NewReader(testRSS("go", "https://example.com/go")+testRSS("dog", "https://example.com/dog")),
		&stdout,
		&stderr,
		false,
	)
	if code != 0 {
		t.Fatalf("run exit code = %d, stderr:\n%s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, `"schema_version":"feed2cli.explain.v1"`) {
		t.Fatalf("missing explain schema:\n%s", got)
	}
	if !strings.Contains(got, `"kept":true`) || !strings.Contains(got, `"kept":false`) {
		t.Fatalf("expected kept and dropped records:\n%s", got)
	}
	if !strings.Contains(got, `dropped by exclude=dog`) {
		t.Fatalf("missing drop reason:\n%s", got)
	}
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
