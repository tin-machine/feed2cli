package main

import (
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

func itemDedupKey(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	if key := normalizeFeedURL(item.Link); key != "" {
		return key
	}
	return item.Link
}

func feedItemDedupKey(item FeedItem) string {
	if item.NormalizedURL != "" {
		return item.NormalizedURL
	}
	if key := normalizeFeedURL(item.URL); key != "" {
		return key
	}
	return item.URL
}

func normalizeFeedURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	u.RawQuery = normalizedQuery(u.Query())

	if strings.HasPrefix(u.Host, "m.") {
		u.Host = strings.TrimPrefix(u.Host, "m.")
	}

	u.Path = strings.TrimSuffix(u.Path, "/amp/")
	u.Path = strings.TrimSuffix(u.Path, "/amp")
	if u.Path != "/" {
		u.Path = strings.TrimRight(u.Path, "/")
	}

	return u.String()
}

func normalizedQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	for key := range values {
		lowerKey := strings.ToLower(key)
		if lowerKey == "fbclid" || lowerKey == "gclid" || strings.HasPrefix(lowerKey, "utm_") {
			delete(values, key)
		}
	}
	return values.Encode()
}
