package main

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type KeywordFilterStage struct {
	Include  []string
	Exclude  []string
	MinScore int
}

func (KeywordFilterStage) Name() string {
	return "keyword_filter"
}

func (s KeywordFilterStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		text := strings.ToLower(strings.Join([]string{item.Title, item.Description, item.Content, strings.Join(item.Categories, " ")}, " "))
		if containsAny(text, s.Exclude) {
			continue
		}
		score := keywordScore(text, s.Include)
		if len(s.Include) > 0 && score == 0 {
			continue
		}
		if s.MinScore > 0 && score < s.MinScore {
			continue
		}
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["keyword_score"] = strconv.Itoa(score)
		out = append(out, item)
	}
	return out, nil
}

type DomainFilterStage struct {
	Include []string
	Exclude []string
}

func (DomainFilterStage) Name() string {
	return "domain_filter"
}

func (s DomainFilterStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	include := normalizeDomainList(s.Include)
	exclude := normalizeDomainList(s.Exclude)
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		domain := itemDomain(item)
		if domain == "" {
			if len(include) == 0 {
				out = append(out, item)
			}
			continue
		}
		if domainMatches(domain, exclude) {
			continue
		}
		if len(include) > 0 && !domainMatches(domain, include) {
			continue
		}
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["domain"] = domain
		out = append(out, item)
	}
	return out, nil
}

type TimeWindowFilterStage struct {
	Since time.Duration
	Now   time.Time
}

func (TimeWindowFilterStage) Name() string {
	return "time_window_filter"
}

func (s TimeWindowFilterStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Since <= 0 {
		return append([]FeedItem(nil), items...), nil
	}
	now := s.Now
	if now.IsZero() {
		now = time.Now()
	}
	cutoff := now.Add(-s.Since)
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		published, ok := item.PublishedTime()
		if !ok || published.Before(cutoff) {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

type HotnessScoreStage struct {
	Now time.Time
}

func (HotnessScoreStage) Name() string {
	return "hotness_score"
}

func (s HotnessScoreStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := s.Now
	if now.IsZero() {
		now = time.Now()
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		score := hotnessScore(item, now)
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["hotness_score"] = formatScore(score)
		out = append(out, item)
	}
	return out, nil
}

type MinHotnessStage struct {
	Min float64
}

func (MinHotnessStage) Name() string {
	return "min_hotness"
}

func (s MinHotnessStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Min <= 0 {
		return append([]FeedItem(nil), items...), nil
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		if parseMetadataScore(item, "hotness_score") >= s.Min {
			out = append(out, item)
		}
	}
	return out, nil
}

type FavUserStage struct {
	Users []string
}

func (FavUserStage) Name() string {
	return "fav_user"
}

func (s FavUserStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	watch := make(map[string]struct{}, len(s.Users))
	for _, user := range s.Users {
		user = strings.TrimSpace(strings.ToLower(user))
		if user != "" {
			watch[user] = struct{}{}
		}
	}
	if len(watch) == 0 {
		return append([]FeedItem(nil), items...), nil
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		var matched []string
		for _, comment := range item.HatenaBookmarkComments {
			if _, ok := watch[strings.ToLower(comment.User)]; ok {
				matched = append(matched, comment.User)
			}
		}
		if len(matched) == 0 {
			continue
		}
		item.Metadata = cloneMetadata(item.Metadata)
		item.Metadata["fav_users"] = strings.Join(uniqueStrings(matched), ",")
		out = append(out, item)
	}
	return out, nil
}

type RankStage struct {
	By string
}

func (RankStage) Name() string {
	return "rank"
}

func (s RankStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := append([]FeedItem(nil), items...)
	switch strings.ToLower(strings.TrimSpace(s.By)) {
	case "", "none":
		return out, nil
	case "hotness":
		sort.SliceStable(out, func(i, j int) bool {
			return parseMetadataScore(out[i], "hotness_score") > parseMetadataScore(out[j], "hotness_score")
		})
	case "published", "time":
		SortFeedItems(out)
	default:
		return nil, fmt.Errorf("unsupported rank %q", s.By)
	}
	return out, nil
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern != "" && strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func keywordScore(text string, patterns []string) int {
	score := 0
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern != "" && strings.Contains(text, pattern) {
			score++
		}
	}
	return score
}

func normalizeDomainList(domains []string) []string {
	out := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		domain = strings.TrimPrefix(domain, "www.")
		if domain != "" {
			out = append(out, domain)
		}
	}
	return out
}

func itemDomain(item FeedItem) string {
	raw := item.NormalizedURL
	if raw == "" {
		raw = normalizeFeedURL(item.URL)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	return strings.TrimPrefix(host, "www.")
}

func domainMatches(domain string, patterns []string) bool {
	for _, pattern := range patterns {
		if domain == pattern || strings.HasSuffix(domain, "."+pattern) {
			return true
		}
	}
	return false
}

func hotnessScore(item FeedItem, now time.Time) float64 {
	bookmarks, _ := strconv.Atoi(item.HatenaBookmarkCount)
	comments := len(item.HatenaBookmarkComments)
	ageHours := 24.0
	if published, ok := item.PublishedTime(); ok {
		ageHours = math.Max(now.Sub(published).Hours(), 0)
	}
	freshness := 24.0 / (ageHours + 24.0)
	return float64(bookmarks)*1.5 + float64(comments)*2.0 + freshness*10.0
}

func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'f', 3, 64)
}

func parseMetadataScore(item FeedItem, key string) float64 {
	if item.Metadata == nil {
		return 0
	}
	score, err := strconv.ParseFloat(item.Metadata[key], 64)
	if err != nil {
		return 0
	}
	return score
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}
