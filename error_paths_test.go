package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

type failingWriter struct{}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestOutputJSONToReturnsWriterError(t *testing.T) {
	err := OutputJSONTo(failingWriter{}, []*gofeed.Feed{{Items: []*gofeed.Item{{Title: "item"}}}})
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("OutputJSONTo error = %v", err)
	}
}

func TestOutputStandardToSkipsNilItems(t *testing.T) {
	var out bytes.Buffer
	err := OutputStandardTo(&out, []*FilteredItem{
		nil,
		{Item: nil},
		{Item: &gofeed.Item{Title: "ok", Link: "https://example.com/ok"}},
	}, time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("OutputStandardTo returned error: %v", err)
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("output missing non-nil item:\n%s", out.String())
	}
}

func TestOpenOrCreateFileWithDirsReturnsMkdirError(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(parentFile, []byte("x"), 0666); err != nil {
		t.Fatalf("failed to write parent file: %v", err)
	}
	_, err := openOrCreateFileWithDirs(filepath.Join(parentFile, "child"), parentFile, 0666)
	if err == nil {
		t.Fatal("openOrCreateFileWithDirs returned nil error")
	}
}

func TestFileHatenaStateStoreBrokenJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("{broken"), 0666); err != nil {
		t.Fatalf("failed to write broken state: %v", err)
	}
	_, err := fileHatenaStateStore{path: path}.Load()
	if err == nil {
		t.Fatal("Load returned nil error for broken JSON")
	}
}

func TestRunSlackReturnsEnvError(t *testing.T) {
	t.Setenv("FEED2CLI_SLACK_TOKEN", "")
	t.Setenv("FEED2CLI_SLACK_CHANNEL", "")
	t.Setenv("XOXB", "")
	t.Setenv("SLACK_CHANNEL", "")

	var stdout, stderr bytes.Buffer
	code := run(
		[]string{"feed2cli", "-o", "slack"},
		strings.NewReader(testRSS("slack", "https://example.com/slack")),
		&stdout,
		&stderr,
		false,
	)
	if code == 0 {
		t.Fatalf("run slack exit code = 0, stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "Slack token") {
		t.Fatalf("stderr does not contain env error:\n%s", stderr.String())
	}
}
