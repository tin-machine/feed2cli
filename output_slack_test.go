package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

type fakeSlackPoster struct {
	timestamps []string
	errOnCall  map[int]error
	calls      []fakeSlackCall
}

type fakeSlackCall struct {
	channel string
	options int
}

func (f *fakeSlackPoster) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	callNumber := len(f.calls) + 1
	f.calls = append(f.calls, fakeSlackCall{channel: channelID, options: len(options)})
	if err := f.errOnCall[callNumber]; err != nil {
		return "", "", err
	}
	if len(f.timestamps) >= callNumber {
		return channelID, f.timestamps[callNumber-1], nil
	}
	return channelID, "ts", nil
}

func TestOutputSlackPostsEachItem(t *testing.T) {
	poster := &fakeSlackPoster{}
	var sleeps []time.Duration
	sleep := func(d time.Duration) {
		sleeps = append(sleeps, d)
	}
	feeds := []*gofeed.Feed{
		{Items: []*gofeed.Item{
			{Title: "one", Link: "https://example.com/one"},
			{Title: "two", Link: "https://example.com/two"},
		}},
	}

	outputSlack(feeds, poster, "C123", sleep)

	if len(poster.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(poster.calls))
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %d, want 2", len(sleeps))
	}
	for _, call := range poster.calls {
		if call.channel != "C123" {
			t.Fatalf("channel = %q, want C123", call.channel)
		}
	}
}

func TestOutputSlackContinuesAfterPostError(t *testing.T) {
	poster := &fakeSlackPoster{errOnCall: map[int]error{1: errors.New("slack down")}}
	feeds := []*gofeed.Feed{
		{Items: []*gofeed.Item{
			{Title: "one", Link: "https://example.com/one"},
			{Title: "two", Link: "https://example.com/two"},
		}},
	}

	outputSlack(feeds, poster, "C123", func(time.Duration) {})

	if len(poster.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(poster.calls))
	}
}

func TestFormatSlackAttachment(t *testing.T) {
	attachment := formatSlackAttachment(&FilteredItem{
		Item:                &gofeed.Item{Description: "description"},
		HatenaBookmarkCount: "3",
		HatenaBookmarkComments: []HatenaBookmarkComment{
			{User: "alice", Comment: "nice"},
			{User: "bob", Comment: "good"},
		},
	})

	for _, want := range []string{"description", "Hatena Bookmark: *3*", "Hatena Comments: 2"} {
		if !strings.Contains(attachment.Text, want) {
			t.Fatalf("attachment text does not contain %q: %q", want, attachment.Text)
		}
	}
}
