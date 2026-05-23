package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const feedItemJSONLSchemaVersion = "feed2cli.feed_item.v1"

type FeedItemJSONLRecord struct {
	SchemaVersion string `json:"schema_version,omitempty"`
	FeedItem
}

func NewFeedItemJSONLRecord(item FeedItem) FeedItemJSONLRecord {
	if item.NormalizedURL == "" {
		item.NormalizedURL = normalizeFeedURL(item.URL)
	}
	if item.ID == "" {
		item.ID = feedItemDedupKey(item)
	}
	return FeedItemJSONLRecord{
		SchemaVersion: feedItemJSONLSchemaVersion,
		FeedItem:      item,
	}
}

func ReadJSONLItems(r io.Reader) ([]FeedItem, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var items []FeedItem
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		item, err := decodeFeedItemJSONLLine([]byte(line))
		if err != nil {
			return nil, fmt.Errorf("jsonl line %d: %w", lineNumber, err)
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func decodeFeedItemJSONLLine(data []byte) (FeedItem, error) {
	var record FeedItemJSONLRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return FeedItem{}, err
	}
	if record.SchemaVersion != "" && record.SchemaVersion != feedItemJSONLSchemaVersion {
		return FeedItem{}, fmt.Errorf("unsupported schema_version %q", record.SchemaVersion)
	}
	item := record.FeedItem
	if item.NormalizedURL == "" {
		item.NormalizedURL = normalizeFeedURL(item.URL)
	}
	if item.ID == "" {
		item.ID = feedItemDedupKey(item)
	}
	return item, nil
}
