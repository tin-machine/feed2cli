package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestLintFeedsFromReaderValidAndInvalid(t *testing.T) {
	input := testRSS("ok", "https://example.com/ok") + `not xml</rss>`
	result := LintFeedsFromReader(strings.NewReader(input))
	if result.Total != 2 || result.Valid != 1 || result.Invalid != 1 {
		t.Fatalf("lint result = %#v", result)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("lint errors = %#v", result.Errors)
	}
}

func TestOutputFeedLintTo(t *testing.T) {
	var out bytes.Buffer
	result := FeedLintResult{Total: 2, Valid: 1, Invalid: 1, Errors: []string{"feed 2: broken"}}
	if err := OutputFeedLintTo(&out, result); err != nil {
		t.Fatalf("OutputFeedLintTo returned error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "feeds: total=2 valid=1 invalid=1") || !strings.Contains(got, "feed 2: broken") {
		t.Fatalf("unexpected lint output:\n%s", got)
	}
}

func TestRunLintReturnsNonZeroForInvalidFeed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"feed2cli", "-o", "lint"}, strings.NewReader(`not xml</rss>`), &stdout, &stderr, false)
	if code == 0 {
		t.Fatalf("run lint exit code = 0, stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "invalid=1") {
		t.Fatalf("lint stdout missing invalid count:\n%s", stdout.String())
	}
}
