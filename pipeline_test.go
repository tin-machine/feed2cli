package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunFeedItemStagesNormalizeAndMerge(t *testing.T) {
	items := []FeedItem{
		{Title: "raw", URL: "https://example.com/post?utm_source=x"},
		{Title: "canonical", URL: "https://example.com/post"},
	}

	got, err := RunFeedItemStages(context.Background(), items, NormalizeStage{}, MergeStage{})
	if err != nil {
		t.Fatalf("RunFeedItemStages returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].NormalizedURL != "https://example.com/post" {
		t.Fatalf("NormalizedURL = %q", got[0].NormalizedURL)
	}
}

func TestRunFeedItemStagesStopsOnError(t *testing.T) {
	wantErr := errors.New("stage failed")
	_, err := RunFeedItemStages(
		context.Background(),
		[]FeedItem{{Title: "item"}},
		FeedItemStageFunc{
			StageName: "fail",
			Fn: func(context.Context, []FeedItem) ([]FeedItem, error) {
				return nil, wantErr
			},
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestRunFeedItemStagesHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := RunFeedItemStages(ctx, []FeedItem{{Title: "item"}}, NormalizeStage{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestDiffStage(t *testing.T) {
	got, err := DiffStage{
		Existing: []FeedItem{{URL: "https://example.com/new", NormalizedURL: "https://example.com/new"}},
	}.Apply(context.Background(), []FeedItem{
		{Title: "old", URL: "https://example.com/old", NormalizedURL: "https://example.com/old"},
		{Title: "new", URL: "https://example.com/new", NormalizedURL: "https://example.com/new"},
	})
	if err != nil {
		t.Fatalf("DiffStage returned error: %v", err)
	}
	if len(got) != 1 || got[0].Title != "old" {
		t.Fatalf("diff result = %#v", got)
	}
}

func TestHatenaBookmarkStage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/count":
			_, _ = w.Write([]byte("7"))
		case "/comments":
			_, _ = w.Write([]byte(`{"bookmarks":[{"user":"alice","comment":"nice","timestamp":"2026/05/23 12:00"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stage := HatenaBookmarkStage{Filter: &HatenaBookmarkFilter{
		Client:           server.Client(),
		CountEndpoint:    server.URL + "/count",
		CommentsEndpoint: server.URL + "/comments",
	}}

	got, err := stage.Apply(context.Background(), []FeedItem{{
		Title:         "item",
		URL:           "https://example.com/item",
		NormalizedURL: "https://example.com/item",
		Source:        "source",
		Metadata:      map[string]string{"k": "v"},
	}})
	if err != nil {
		t.Fatalf("HatenaBookmarkStage returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d", len(got))
	}
	if got[0].HatenaBookmarkCount != "7" || len(got[0].HatenaBookmarkComments) != 1 {
		t.Fatalf("hatena enrichment missing: %#v", got[0])
	}
	if got[0].Source != "source" || got[0].Metadata["k"] != "v" {
		t.Fatalf("source/metadata not preserved: %#v", got[0])
	}
}
