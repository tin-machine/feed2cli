package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HatenaBookmarkComment は、はてなブックマークのコメント一件を表す構造体です
type HatenaBookmarkComment struct {
	User      string   `json:"user"`
	Comment   string   `json:"comment"`
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
}

// getHatenaBookmarkComments は、指定されたURLのはてなブックマークコメントを取得します
func GetHatenaBookmarkComments(entryURL string) ([]HatenaBookmarkComment, error) {
	apiURL := fmt.Sprintf("http://b.hatena.ne.jp/entry/jsonlite/?url=%s", entryURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch comments: status code %d", resp.StatusCode)
	}

	var data struct {
		Bookmarks []HatenaBookmarkComment `json:"bookmarks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Bookmarks, nil
}
