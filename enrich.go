package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SourceLabelStage struct {
	DefaultSource string
}

func (SourceLabelStage) Name() string {
	return "source_label"
}

func (s SourceLabelStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		enriched := item
		source := strings.TrimSpace(enriched.Source)
		if source == "" {
			source = s.DefaultSource
		}
		if source == "" && enriched.Raw != nil && enriched.Raw.Author != nil {
			source = enriched.Raw.Author.Name
		}
		enriched.Source = source
		enriched.Metadata = cloneMetadata(enriched.Metadata)
		if source != "" {
			enriched.Metadata["source_label"] = source
		}
		out = append(out, enriched)
	}
	return out, nil
}

type TagEnrichStage struct{}

func (TagEnrichStage) Name() string {
	return "tag"
}

func (TagEnrichStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		enriched := item
		enriched.Categories = mergeTags(enriched.Categories, hatenaCommentTags(enriched.HatenaBookmarkComments))
		out = append(out, enriched)
	}
	return out, nil
}

type OGPEnrichStage struct {
	Client *http.Client
}

func (OGPEnrichStage) Name() string {
	return "ogp"
}

func (s OGPEnrichStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	client := s.httpClient()
	for _, item := range items {
		enriched := item
		if item.URL == "" {
			out = append(out, enriched)
			continue
		}
		metadata, err := fetchOGPMetadata(ctx, client, item.URL)
		if err != nil {
			enriched.Metadata = cloneMetadata(enriched.Metadata)
			enriched.Metadata["ogp_error"] = err.Error()
			out = append(out, enriched)
			continue
		}
		enriched.Metadata = cloneMetadata(enriched.Metadata)
		for key, value := range metadata {
			enriched.Metadata[key] = value
		}
		if enriched.Title == "" {
			enriched.Title = firstNonEmpty(metadata["og:title"], metadata["title"])
		}
		if enriched.Description == "" {
			enriched.Description = firstNonEmpty(metadata["og:description"], metadata["description"])
		}
		out = append(out, enriched)
	}
	return out, nil
}

func (s OGPEnrichStage) httpClient() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

type ContentFetchStage struct {
	Client *http.Client
}

func (ContentFetchStage) Name() string {
	return "content"
}

func (s ContentFetchStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	client := s.httpClient()
	for _, item := range items {
		enriched := item
		if item.URL == "" {
			out = append(out, enriched)
			continue
		}
		content, err := fetchReadableContent(ctx, client, item.URL)
		enriched.Metadata = cloneMetadata(enriched.Metadata)
		if err != nil {
			enriched.Metadata["content_error"] = err.Error()
			out = append(out, enriched)
			continue
		}
		enriched.Content = content
		enriched.Metadata["content_text"] = content
		out = append(out, enriched)
	}
	return out, nil
}

func (s ContentFetchStage) httpClient() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

type SummaryStage struct {
	Summarizer Summarizer
}

func (SummaryStage) Name() string {
	return "summary"
}

func (s SummaryStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	summarizer := s.Summarizer
	if summarizer == nil {
		summarizer = ExtractiveSummarizer{MaxRunes: 160}
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		summary, reason, err := summarizer.Summarize(ctx, item)
		enriched := item
		enriched.Metadata = cloneMetadata(enriched.Metadata)
		if err != nil {
			enriched.Metadata["summary_error"] = err.Error()
			out = append(out, enriched)
			continue
		}
		enriched.Summary = summary
		enriched.ReadingReason = reason
		if summary != "" {
			enriched.Metadata["summary"] = summary
		}
		if reason != "" {
			enriched.Metadata["reading_reason"] = reason
		}
		out = append(out, enriched)
	}
	return out, nil
}

type Summarizer interface {
	Summarize(context.Context, FeedItem) (summary, reason string, err error)
}

type ExtractiveSummarizer struct {
	MaxRunes int
}

func (s ExtractiveSummarizer) Summarize(ctx context.Context, item FeedItem) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	maxRunes := s.MaxRunes
	if maxRunes <= 0 {
		maxRunes = 160
	}
	source := firstNonEmpty(item.Content, item.Description, item.Title)
	summary := truncateRunes(strings.Join(strings.Fields(source), " "), maxRunes)
	reason := "source feed item"
	if item.HatenaBookmarkCount != "" && item.HatenaBookmarkCount != "0" {
		reason = fmt.Sprintf("Hatena Bookmark: %s", item.HatenaBookmarkCount)
	} else if item.Source != "" {
		reason = fmt.Sprintf("source: %s", item.Source)
	}
	return summary, reason, nil
}

func fetchOGPMetadata(ctx context.Context, client *http.Client, itemURL string) (map[string]string, error) {
	doc, err := fetchHTMLDocument(ctx, client, itemURL)
	if err != nil {
		return nil, err
	}
	metadata := map[string]string{}
	doc.Find("meta").Each(func(_ int, selection *goquery.Selection) {
		key, _ := selection.Attr("property")
		if key == "" {
			key, _ = selection.Attr("name")
		}
		value, _ := selection.Attr("content")
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return
		}
		switch key {
		case "og:title", "og:description", "og:image", "og:site_name", "description":
			metadata[key] = value
		}
	})
	if title := strings.TrimSpace(doc.Find("title").First().Text()); title != "" {
		metadata["title"] = title
	}
	return metadata, nil
}

func fetchReadableContent(ctx context.Context, client *http.Client, itemURL string) (string, error) {
	doc, err := fetchHTMLDocument(ctx, client, itemURL)
	if err != nil {
		return "", err
	}
	doc.Find("script, style, nav, footer, header, noscript").Remove()
	for _, selector := range []string{"article", "main", "body"} {
		text := cleanText(doc.Find(selector).First().Text())
		if text != "" {
			return text, nil
		}
	}
	return "", nil
}

func fetchHTMLDocument(ctx context.Context, client *http.Client, itemURL string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, itemURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, 2*1024*1024)
	return goquery.NewDocumentFromReader(limited)
}

func cloneMetadata(metadata map[string]string) map[string]string {
	cloned := make(map[string]string, len(metadata)+4)
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func hatenaCommentTags(comments []HatenaBookmarkComment) []string {
	var tags []string
	for _, comment := range comments {
		tags = append(tags, comment.Tags...)
	}
	return tags
}

func mergeTags(existing, additions []string) []string {
	seen := map[string]struct{}{}
	var merged []string
	for _, tag := range append(append([]string{}, existing...), additions...) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		merged = append(merged, tag)
	}
	return merged
}

func cleanText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truncateRunes(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
