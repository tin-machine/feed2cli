package main

import (
	"strings"
	"testing"
)

func TestSplitFeed(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		atEOF       bool
		wantAdvance int
		wantToken   string
	}{
		{
			name:        "rss",
			data:        `<rss><channel></channel></rss><rss></rss>`,
			wantAdvance: len(`<rss><channel></channel></rss>`),
			wantToken:   `<rss><channel></channel></rss>`,
		},
		{
			name:        "atom",
			data:        `<feed><entry></entry></feed><feed></feed>`,
			wantAdvance: len(`<feed><entry></entry></feed>`),
			wantToken:   `<feed><entry></entry></feed>`,
		},
		{
			name:        "eof returns remaining data",
			data:        `<rss>unfinished`,
			atEOF:       true,
			wantAdvance: len(`<rss>unfinished`),
			wantToken:   `<rss>unfinished`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advance, token, err := splitFeed([]byte(tt.data), tt.atEOF)
			if err != nil {
				t.Fatalf("splitFeed returned error: %v", err)
			}
			if advance != tt.wantAdvance {
				t.Fatalf("advance = %d, want %d", advance, tt.wantAdvance)
			}
			if string(token) != tt.wantToken {
				t.Fatalf("token = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestReadFromMultipleFeeds(t *testing.T) {
	input := testRSS("rss one", "https://example.com/1") +
		testAtom("atom one", "https://example.com/2") +
		testRSS("rss two", "https://example.com/3")

	feeds := readFrom(strings.NewReader(input))
	if len(feeds) != 3 {
		t.Fatalf("len(feeds) = %d, want 3", len(feeds))
	}

	wantTitles := []string{"rss one", "atom one", "rss two"}
	for i, want := range wantTitles {
		if feeds[i].Title != want {
			t.Fatalf("feeds[%d].Title = %q, want %q", i, feeds[i].Title, want)
		}
	}
}

func TestReadFromSkipsInvalidFeed(t *testing.T) {
	input := `not xml</rss>` + testRSS("valid", "https://example.com/valid")

	feeds := readFrom(strings.NewReader(input))
	if len(feeds) != 1 {
		t.Fatalf("len(feeds) = %d, want 1", len(feeds))
	}
	if feeds[0].Title != "valid" {
		t.Fatalf("feeds[0].Title = %q, want valid", feeds[0].Title)
	}
}

func testRSS(title, link string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>` + title + `</title>
    <link>` + link + `</link>
    <description>` + title + ` description</description>
    <item>
      <title>` + title + ` item</title>
      <link>` + link + `/item</link>
      <guid>` + link + `/item</guid>
      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>
    </item>
  </channel>
</rss>`
}

func testAtom(title, link string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>` + title + `</title>
  <id>` + link + `</id>
  <updated>2006-01-02T15:04:05Z</updated>
  <entry>
    <title>` + title + ` item</title>
    <id>` + link + `/item</id>
    <link href="` + link + `/item"/>
    <updated>2006-01-02T15:04:05Z</updated>
  </entry>
</feed>`
}
