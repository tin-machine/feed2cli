package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestCLIIntegration(t *testing.T) {
	binary := buildTestBinary(t)

	t.Run("merge", func(t *testing.T) {
		input := testRSS("old", "https://example.com/old") + testRSS("new", "https://example.com/new")
		output := runCLI(t, binary, []string{"-o", "merge"}, input)

		feed, err := gofeed.NewParser().ParseString(output)
		if err != nil {
			t.Fatalf("merge output did not parse: %v\n%s", err, output)
		}
		if len(feed.Items) != 2 {
			t.Fatalf("merge item count = %d, want 2", len(feed.Items))
		}
	})

	t.Run("diff", func(t *testing.T) {
		oldFeed := testRSS("old", "https://example.com/old")
		newFeed := testRSS("new", "https://example.com/new")
		output := runCLI(t, binary, []string{"-o", "diff"}, oldFeed+newFeed)

		feed, err := gofeed.NewParser().ParseString(output)
		if err != nil {
			t.Fatalf("diff output did not parse: %v\n%s", err, output)
		}
		if len(feed.Items) != 1 {
			t.Fatalf("diff item count = %d, want 1", len(feed.Items))
		}
		if !strings.Contains(feed.Items[0].Link, "old") {
			t.Fatalf("diff item link = %q, want old item", feed.Items[0].Link)
		}
	})

	t.Run("json", func(t *testing.T) {
		output := runCLI(t, binary, []string{"-o", "json"}, testRSS("json", "https://example.com/json"))
		if !strings.Contains(output, `"title": "json item"`) {
			t.Fatalf("json output does not contain item title:\n%s", output)
		}
	})

	t.Run("symlink mergeRss", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "mergeRss")
		if err := os.Symlink(binary, link); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}
		output := runCLI(t, link, nil, testRSS("one", "https://example.com/one")+testRSS("two", "https://example.com/two"))
		feed, err := gofeed.NewParser().ParseString(output)
		if err != nil {
			t.Fatalf("mergeRss output did not parse: %v\n%s", err, output)
		}
		if len(feed.Items) != 2 {
			t.Fatalf("mergeRss item count = %d, want 2", len(feed.Items))
		}
	})
}

func buildTestBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "feed2cli")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}

func runCLI(t *testing.T, binary string, args []string, input string) string {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", binary, args, err, output)
	}
	return string(output)
}
