package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestHatenaBookmarkFilterApply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/count":
			if got := r.URL.Query().Get("url"); got != "https://example.com/item?x=1&y=2" {
				t.Fatalf("count url query = %q", got)
			}
			fmt.Fprint(w, "42")
		case "/comments":
			if got := r.URL.Query().Get("url"); got != "https://example.com/item?x=1&y=2" {
				t.Fatalf("comments url query = %q", got)
			}
			fmt.Fprint(w, `{"bookmarks":[{"user":"alice","comment":"nice","timestamp":"2026/05/20 12:00","tags":["go"]}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	filter := &HatenaBookmarkFilter{
		Client:           server.Client(),
		CountEndpoint:    server.URL + "/count",
		CommentsEndpoint: server.URL + "/comments",
	}

	items, err := filter.Apply([]*gofeed.Item{{Title: "item", Link: "https://example.com/item?x=1&y=2"}})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].HatenaBookmarkCount != "42" {
		t.Fatalf("HatenaBookmarkCount = %q, want 42", items[0].HatenaBookmarkCount)
	}
	if len(items[0].HatenaBookmarkComments) != 1 {
		t.Fatalf("len(HatenaBookmarkComments) = %d, want 1", len(items[0].HatenaBookmarkComments))
	}
	if items[0].HatenaBookmarkComments[0].User != "alice" {
		t.Fatalf("comment user = %q, want alice", items[0].HatenaBookmarkComments[0].User)
	}
}

func TestHatenaBookmarkFilterApplyNilAndEmptyLink(t *testing.T) {
	filter := &HatenaBookmarkFilter{}

	items, err := filter.Apply([]*gofeed.Item{nil, &gofeed.Item{Title: "no link"}})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Item != nil {
		t.Fatalf("items[0].Item = %v, want nil", items[0].Item)
	}
	if items[1].HatenaBookmarkCount != "" {
		t.Fatalf("empty link count = %q, want empty", items[1].HatenaBookmarkCount)
	}
}

func TestGetHatenaBookmarkCount(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
		wantErr    bool
	}{
		{name: "ok", statusCode: http.StatusOK, body: "7\n", want: "7"},
		{name: "empty count is zero", statusCode: http.StatusOK, want: "0"},
		{name: "server error", statusCode: http.StatusInternalServerError, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			filter := &HatenaBookmarkFilter{Client: server.Client(), CountEndpoint: server.URL}
			got, err := filter.getHatenaBookmarkCount("https://example.com/item")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("getHatenaBookmarkCount returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("count = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetHatenaBookmarkComments(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantLen    int
		wantErr    bool
	}{
		{name: "ok", statusCode: http.StatusOK, body: `{"bookmarks":[{"user":"alice","comment":"nice","timestamp":"2026/05/20 12:00"}]}`, wantLen: 1},
		{name: "not found", statusCode: http.StatusNotFound, wantLen: 0},
		{name: "empty body", statusCode: http.StatusOK, wantLen: 0},
		{name: "broken json", statusCode: http.StatusOK, body: `{`, wantErr: true},
		{name: "server error", statusCode: http.StatusInternalServerError, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			filter := &HatenaBookmarkFilter{Client: server.Client(), CommentsEndpoint: server.URL}
			got, err := filter.getHatenaBookmarkComments("https://example.com/item")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("getHatenaBookmarkComments returned error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len(comments) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
