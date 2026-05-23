package main

import (
	"context"

	"github.com/mmcdole/gofeed"
)

type FeedItemStage interface {
	Name() string
	Apply(context.Context, []FeedItem) ([]FeedItem, error)
}

type FeedItemStageFunc struct {
	StageName string
	Fn        func(context.Context, []FeedItem) ([]FeedItem, error)
}

func (s FeedItemStageFunc) Name() string {
	return s.StageName
}

func (s FeedItemStageFunc) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	return s.Fn(ctx, items)
}

func RunFeedItemStages(ctx context.Context, items []FeedItem, stages ...FeedItemStage) ([]FeedItem, error) {
	current := append([]FeedItem(nil), items...)
	for _, stage := range stages {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		next, err := stage.Apply(ctx, current)
		if err != nil {
			return nil, err
		}
		current = next
	}
	return current, nil
}

type NormalizeStage struct{}

func (NormalizeStage) Name() string {
	return "normalize"
}

func (NormalizeStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]FeedItem, 0, len(items))
	for _, item := range items {
		normalized := item
		if normalized.NormalizedURL == "" {
			normalized.NormalizedURL = normalizeFeedURL(normalized.URL)
		}
		if normalized.ID == "" {
			normalized.ID = feedItemDedupKey(normalized)
		}
		out = append(out, normalized)
	}
	return out, nil
}

type MergeStage struct{}

func (MergeStage) Name() string {
	return "merge"
}

func (MergeStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return MergeFeedItems(items), nil
}

type DiffStage struct {
	Existing []FeedItem
}

func (DiffStage) Name() string {
	return "diff"
}

func (s DiffStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return DiffFeedItems(items, s.Existing), nil
}

type HatenaBookmarkStage struct {
	Filter *HatenaBookmarkFilter
}

func (HatenaBookmarkStage) Name() string {
	return "hatena_bookmark"
}

func (s HatenaBookmarkStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	filter := s.Filter
	if filter == nil {
		filter = &HatenaBookmarkFilter{}
	}
	gofeedItems := make([]*gofeed.Item, 0, len(items))
	for _, item := range items {
		gofeedItems = append(gofeedItems, item.ToGofeedItem())
	}
	filtered, err := filter.Apply(gofeedItems)
	if err != nil {
		return nil, err
	}
	enriched := FeedItemsFromFilteredItems(filtered)
	for i := range enriched {
		if i < len(items) {
			enriched[i].Source = items[i].Source
			enriched[i].Metadata = items[i].Metadata
		}
	}
	return enriched, nil
}
