package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/mmcdole/gofeed"
)

type FeedLintResult struct {
	Total   int
	Valid   int
	Invalid int
	Errors  []string
}

func LintFeedsFromReader(r io.Reader) FeedLintResult {
	fp := gofeed.NewParser()
	scanner := bufio.NewScanner(r)
	buf := make([]byte, initialFeedScannerBufferSize)
	scanner.Buffer(buf, maxFeedScannerBufferSize)
	scanner.Split(splitFeed)

	var result FeedLintResult
	for scanner.Scan() {
		xmlData := strings.Map(printOnly, string(scanner.Text()))
		if strings.TrimSpace(xmlData) == "" {
			continue
		}
		result.Total++
		if _, err := fp.ParseString(xmlData); err != nil {
			result.Invalid++
			result.Errors = append(result.Errors, fmt.Sprintf("feed %d: %v", result.Total, err))
			continue
		}
		result.Valid++
	}

	if err := scanner.Err(); err != nil {
		result.Invalid++
		result.Errors = append(result.Errors, fmt.Sprintf("scanner: %v", err))
	}
	return result
}

func LintFeedURLs(feedURLs []string) FeedLintResult {
	var result FeedLintResult
	for _, feedURL := range feedURLs {
		result.Total++
		if _, err := fetchFeedFromURL(feedURL); err != nil {
			result.Invalid++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", feedURL, err))
			continue
		}
		result.Valid++
	}
	return result
}

func MergeFeedLintResults(results ...FeedLintResult) FeedLintResult {
	var merged FeedLintResult
	for _, result := range results {
		merged.Total += result.Total
		merged.Valid += result.Valid
		merged.Invalid += result.Invalid
		merged.Errors = append(merged.Errors, result.Errors...)
	}
	return merged
}

func OutputFeedLintTo(w io.Writer, result FeedLintResult) error {
	if _, err := fmt.Fprintf(w, "feeds: total=%d valid=%d invalid=%d\n", result.Total, result.Valid, result.Invalid); err != nil {
		return err
	}
	for _, lintErr := range result.Errors {
		if _, err := fmt.Fprintf(w, "- %s\n", lintErr); err != nil {
			return err
		}
	}
	return nil
}
