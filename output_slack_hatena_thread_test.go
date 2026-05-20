package main

import (
	"errors"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

type memoryHatenaStateStore struct {
	state     State
	loadErr   error
	saveErr   error
	saveCount int
}

func (s *memoryHatenaStateStore) Load() (State, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	if s.state == nil {
		return make(State), nil
	}
	copied := make(State, len(s.state))
	for key, value := range s.state {
		copied[key] = value
	}
	return copied, nil
}

func (s *memoryHatenaStateStore) Save(state State) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saveCount++
	s.state = make(State, len(state))
	for key, value := range state {
		s.state[key] = value
	}
	return nil
}

func TestOutputHatenaToSlackNewEntryPostsParentAndComments(t *testing.T) {
	poster := &fakeSlackPoster{timestamps: []string{"parent-ts", "comment-1", "comment-2"}}
	store := &memoryHatenaStateStore{}
	var sleeps []time.Duration
	item := filteredHatenaItem("https://example.com/entry", []HatenaBookmarkComment{
		{User: "alice", Comment: "first", Timestamp: "2026/05/20 12:00"},
		{User: "bob", Comment: "second", Timestamp: "2026/05/20 12:01"},
	})

	err := outputHatenaToSlack([]*FilteredItem{item}, poster, "C123", store, func(d time.Duration) {
		sleeps = append(sleeps, d)
	})
	if err != nil {
		t.Fatalf("outputHatenaToSlack returned error: %v", err)
	}
	if len(poster.calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(poster.calls))
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %d, want 2", len(sleeps))
	}
	state := store.state["https://example.com/entry"]
	if state.SlackThreadTimestamp != "parent-ts" {
		t.Fatalf("SlackThreadTimestamp = %q, want parent-ts", state.SlackThreadTimestamp)
	}
	if state.LastCommentTimestamp != "2026-05-20T12:01:00Z" {
		t.Fatalf("LastCommentTimestamp = %q", state.LastCommentTimestamp)
	}
}

func TestOutputHatenaToSlackExistingEntryPostsOnlyNewComments(t *testing.T) {
	poster := &fakeSlackPoster{timestamps: []string{"new-comment"}}
	store := &memoryHatenaStateStore{state: State{
		"https://example.com/entry": {
			SlackThreadTimestamp: "thread-ts",
			LastCommentTimestamp: "2026-05-20T12:00:00Z",
		},
	}}
	item := filteredHatenaItem("https://example.com/entry", []HatenaBookmarkComment{
		{User: "old", Comment: "old", Timestamp: "2026/05/20 11:59"},
		{User: "equal", Comment: "equal", Timestamp: "2026/05/20 12:00"},
		{User: "new", Comment: "new", Timestamp: "2026/05/20 12:01"},
		{User: "bad", Comment: "bad", Timestamp: "not a timestamp"},
	})

	err := outputHatenaToSlack([]*FilteredItem{item}, poster, "C123", store, func(time.Duration) {})
	if err != nil {
		t.Fatalf("outputHatenaToSlack returned error: %v", err)
	}
	if len(poster.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(poster.calls))
	}
	state := store.state["https://example.com/entry"]
	if state.SlackThreadTimestamp != "thread-ts" {
		t.Fatalf("SlackThreadTimestamp = %q, want thread-ts", state.SlackThreadTimestamp)
	}
	if state.LastCommentTimestamp != "2026-05-20T12:01:00Z" {
		t.Fatalf("LastCommentTimestamp = %q", state.LastCommentTimestamp)
	}
}

func TestOutputHatenaToSlackDoesNotAdvanceTimestampWhenCommentPostFails(t *testing.T) {
	poster := &fakeSlackPoster{errOnCall: map[int]error{1: errors.New("slack down")}}
	store := &memoryHatenaStateStore{state: State{
		"https://example.com/entry": {
			SlackThreadTimestamp: "thread-ts",
			LastCommentTimestamp: "2026-05-20T12:00:00Z",
		},
	}}
	item := filteredHatenaItem("https://example.com/entry", []HatenaBookmarkComment{
		{User: "new", Comment: "new", Timestamp: "2026/05/20 12:01"},
	})

	err := outputHatenaToSlack([]*FilteredItem{item}, poster, "C123", store, func(time.Duration) {})
	if err != nil {
		t.Fatalf("outputHatenaToSlack returned error: %v", err)
	}
	if store.saveCount != 0 {
		t.Fatalf("saveCount = %d, want 0", store.saveCount)
	}
	state := store.state["https://example.com/entry"]
	if state.LastCommentTimestamp != "2026-05-20T12:00:00Z" {
		t.Fatalf("LastCommentTimestamp = %q", state.LastCommentTimestamp)
	}
}

func TestOutputHatenaToSlackReturnsStoreErrors(t *testing.T) {
	poster := &fakeSlackPoster{}
	store := &memoryHatenaStateStore{loadErr: errors.New("load failed")}

	err := outputHatenaToSlack(nil, poster, "C123", store, nil)
	if err == nil {
		t.Fatal("expected load error, got nil")
	}

	store = &memoryHatenaStateStore{saveErr: errors.New("save failed")}
	item := filteredHatenaItem("https://example.com/entry", nil)
	err = outputHatenaToSlack([]*FilteredItem{item}, poster, "C123", store, nil)
	if err == nil {
		t.Fatal("expected save error, got nil")
	}
}

func TestFileHatenaStateStore(t *testing.T) {
	store := fileHatenaStateStore{path: t.TempDir() + "/state.json"}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load missing state returned error: %v", err)
	}
	if len(state) != 0 {
		t.Fatalf("len(state) = %d, want 0", len(state))
	}

	want := State{
		"https://example.com/entry": {
			SlackThreadTimestamp: "thread-ts",
			LastCommentTimestamp: "2026-05-20T12:00:00Z",
		},
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got["https://example.com/entry"] != want["https://example.com/entry"] {
		t.Fatalf("loaded state = %#v, want %#v", got, want)
	}
}

func filteredHatenaItem(link string, comments []HatenaBookmarkComment) *FilteredItem {
	return &FilteredItem{
		Item: &gofeed.Item{
			Title:       "entry",
			Link:        link,
			Description: "description",
		},
		HatenaBookmarkComments: comments,
	}
}
