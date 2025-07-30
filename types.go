package main

import "github.com/mmcdole/gofeed"

// HatenaBookmarkComment は、はてなブックマークのコメント一件を表す構造体です。
// filter.go と output_hatena_slack.go の両方から参照されるため、ここに定義します。
type HatenaBookmarkComment struct {
	User      string   `json:"user"`
	Comment   string   `json:"comment"`
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
}

// FilteredItem は、フィルタによって追加情報が付与されたフィードアイテムを表します。
// gofeed.Itemを埋め込むことで、元のフィールドに直接アクセスできます。
type FilteredItem struct {
	*gofeed.Item
	HatenaBookmarkCount    string
	HatenaBookmarkComments []HatenaBookmarkComment
}
