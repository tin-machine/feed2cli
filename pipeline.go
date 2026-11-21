package main

import (
	"context"

	"github.com/mmcdole/gofeed"
)

// FeedItemStage は名前を持ち、入力記事の一覧を処理するパイプラインの一段を表す。
type FeedItemStage interface {
	Name() string
	Apply(context.Context, []FeedItem) ([]FeedItem, error)
}

// FeedItemStageFunc は関数を名前付きの処理ステージとして扱う。
type FeedItemStageFunc struct {
	StageName string
	Fn        func(context.Context, []FeedItem) ([]FeedItem, error)
}

// Name はパイプラインでこのステージを識別する名前を返す。
func (s FeedItemStageFunc) Name() string {
	return s.StageName
}

// Apply は登録した関数にコンテキストと記事を渡す。
func (s FeedItemStageFunc) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	return s.Fn(ctx, items)
}

// RunFeedItemStages は設定順に記事を処理し、キャンセルまたはエラーで終了する。
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

// NormalizeStage は記事の識別情報を補うステージである。
type NormalizeStage struct{}

// Name はパイプラインでこのステージを識別する名前を返す。
func (NormalizeStage) Name() string {
	return "normalize"
}

// Apply は空の正規化 URL と記事 ID を生成する。
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

// MergeStage は重複する記事を統合するステージである。
type MergeStage struct{}

// Name はパイプラインでこのステージを識別する名前を返す。
func (MergeStage) Name() string {
	return "merge"
}

// Apply はURL の重複を取り除き、日時順に並べる。
func (MergeStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return MergeFeedItems(items), nil
}

// DiffStage は比較対象となる既存の記事を保持するステージである。
type DiffStage struct {
	Existing []FeedItem
}

// Name はパイプラインでこのステージを識別する名前を返す。
func (DiffStage) Name() string {
	return "diff"
}

// Apply は既存の一覧に含まれない入力記事を返す。
func (s DiffStage) Apply(ctx context.Context, items []FeedItem) ([]FeedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return DiffFeedItems(items, s.Existing), nil
}

// HatenaBookmarkStage ははてなブックマーク情報を取得するステージである。
type HatenaBookmarkStage struct {
	Filter *HatenaBookmarkFilter
}

// Name はパイプラインでこのステージを識別する名前を返す。
func (HatenaBookmarkStage) Name() string {
	return "hatena_bookmark"
}

// Apply はブックマーク情報を記事に反映して元の出典を引き継ぐ。
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
