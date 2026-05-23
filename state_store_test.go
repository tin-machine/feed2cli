package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSQLiteHatenaStateStoreRoundTrip(t *testing.T) {
	store := sqliteHatenaStateStore{path: filepath.Join(t.TempDir(), "state.sqlite")}
	want := State{
		"https://example.com/entry": {
			LastCommentTimestamp: "2026-05-23T12:00:00Z",
			SlackThreadTimestamp: "123.456",
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded state = %#v, want %#v", got, want)
	}
}

func TestNewHatenaStateStoreSelectsSQLite(t *testing.T) {
	store, err := newHatenaStateStore(hatenaOutputOptions{
		StateBackend: "sqlite",
		StatePath:    filepath.Join(t.TempDir(), "state.sqlite"),
	})
	if err != nil {
		t.Fatalf("newHatenaStateStore returned error: %v", err)
	}
	if _, ok := store.(sqliteHatenaStateStore); !ok {
		t.Fatalf("store type = %T, want sqliteHatenaStateStore", store)
	}
}
