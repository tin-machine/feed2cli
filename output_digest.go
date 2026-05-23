package main

import (
	"fmt"
	"html"
	"io"
	"strings"
	"time"
)

type DigestOptions struct {
	Title  string
	Window time.Duration
	Now    time.Time
}

func OutputDigestMarkdownTo(w io.Writer, data interface{}, options DigestOptions) error {
	if options.Title == "" {
		options.Title = "feed2cli digest"
	}
	if options.Now.IsZero() {
		options.Now = time.Now()
	}

	items := digestItems(FeedItemsFromData(data), options)
	if _, err := fmt.Fprintf(w, "# %s\n\n", options.Title); err != nil {
		return err
	}
	if options.Window > 0 {
		if _, err := fmt.Fprintf(w, "Window: %s - %s\n\n", options.Now.Add(-options.Window).Format(time.RFC3339), options.Now.Format(time.RFC3339)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Items: %d\n\n", len(items)); err != nil {
		return err
	}

	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = item.URL
		}
		if _, err := fmt.Fprintf(w, "- [%s](%s)", escapeMarkdownLinkText(title), item.URL); err != nil {
			return err
		}
		if published, ok := item.PublishedTime(); ok {
			if _, err := fmt.Fprintf(w, " - %s", published.Format(time.RFC3339)); err != nil {
				return err
			}
		}
		if item.HatenaBookmarkCount != "" && item.HatenaBookmarkCount != "0" {
			if _, err := fmt.Fprintf(w, " - Hatena Bookmark: %s", item.HatenaBookmarkCount); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if summary := digestSummary(item.Description); summary != "" {
			if _, err := fmt.Fprintf(w, "  %s\n", summary); err != nil {
				return err
			}
		}
	}
	return nil
}

func digestItems(items []FeedItem, options DigestOptions) []FeedItem {
	cutoff := options.Now.Add(-options.Window)
	filtered := make([]FeedItem, 0, len(items))
	for _, item := range items {
		if options.Window > 0 {
			published, ok := item.PublishedTime()
			if !ok || published.Before(cutoff) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	SortFeedItems(filtered)
	return filtered
}

func digestSummary(description string) string {
	description = strings.TrimSpace(html.UnescapeString(description))
	if description == "" {
		return ""
	}
	description = strings.Join(strings.Fields(description), " ")
	const maxLen = 180
	if len(description) <= maxLen {
		return description
	}
	return description[:maxLen] + "..."
}

func escapeMarkdownLinkText(text string) string {
	text = strings.ReplaceAll(text, "[", `\[`)
	text = strings.ReplaceAll(text, "]", `\]`)
	return text
}
