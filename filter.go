package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Filter は、フィードアイテムのリストを受け取り、
// フィルタリングされたアイテムのリストを返す処理のインターフェースです。
type Filter interface {
	Apply(items []*gofeed.Item) ([]*FilteredItem, error)
}

// HatenaBookmarkFilter は、はてなブックマークの情報を取得し、
// それらをFilteredItemに格納するフィルタです。
type HatenaBookmarkFilter struct {
	Client           *http.Client
	CountEndpoint    string
	CommentsEndpoint string
}

const (
	defaultHatenaCountEndpoint    = "http://api.b.st-hatena.com/entry.count"
	defaultHatenaCommentsEndpoint = "http://b.hatena.ne.jp/entry/jsonlite/"
)

// Apply は、gofeed.Itemのリストを受け取り、はてなブックマーク情報を付与した
// FilteredItemのリストを返します。
func (f *HatenaBookmarkFilter) Apply(items []*gofeed.Item) ([]*FilteredItem, error) {
	filteredItems := make([]*FilteredItem, len(items))

	for i, item := range items {
		if item == nil {
			filteredItems[i] = &FilteredItem{Item: item}
			continue
		}

		var count string
		var comments []HatenaBookmarkComment

		if item.Link != "" {
			count, _ = f.getHatenaBookmarkCount(item.Link)
			comments, _ = f.getHatenaBookmarkComments(item.Link)
		}

		filteredItems[i] = &FilteredItem{
			Item:                   item,
			HatenaBookmarkCount:    count,
			HatenaBookmarkComments: comments,
		}
	}

	return filteredItems, nil
}

// getHatenaBookmarkCount は、指定されたURLのはてなブックマーク数を取得します。
func getHatenaBookmarkCount(entryURL string) (string, error) {
	return (&HatenaBookmarkFilter{}).getHatenaBookmarkCount(entryURL)
}

func (f *HatenaBookmarkFilter) getHatenaBookmarkCount(entryURL string) (string, error) {
	apiURL := buildHatenaURL(f.countEndpoint(), entryURL)
	resp, err := f.httpClient().Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch bookmark count: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	count := strings.TrimSpace(string(body))
	if count == "" {
		return "0", nil
	}
	return count, nil
}

// GetHatenaBookmarkComments は、指定されたURLのはてなブックマークコメントを取得します
func GetHatenaBookmarkComments(entryURL string) ([]HatenaBookmarkComment, error) {
	return (&HatenaBookmarkFilter{}).getHatenaBookmarkComments(entryURL)
}

func (f *HatenaBookmarkFilter) getHatenaBookmarkComments(entryURL string) ([]HatenaBookmarkComment, error) {
	apiURL := buildHatenaURL(f.commentsEndpoint(), entryURL)
	resp, err := f.httpClient().Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return []HatenaBookmarkComment{}, nil
		}
		return nil, fmt.Errorf("failed to fetch comments: status code %d", resp.StatusCode)
	}

	var data struct {
		Bookmarks []HatenaBookmarkComment `json:"bookmarks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		if err.Error() == "EOF" {
			return []HatenaBookmarkComment{}, nil
		}
		return nil, err
	}
	return data.Bookmarks, nil
}

func (f *HatenaBookmarkFilter) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return http.DefaultClient
}

func (f *HatenaBookmarkFilter) countEndpoint() string {
	if f.CountEndpoint != "" {
		return f.CountEndpoint
	}
	return defaultHatenaCountEndpoint
}

func (f *HatenaBookmarkFilter) commentsEndpoint() string {
	if f.CommentsEndpoint != "" {
		return f.CommentsEndpoint
	}
	return defaultHatenaCommentsEndpoint
}

func buildHatenaURL(endpoint, entryURL string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Sprintf("%s?url=%s", endpoint, url.QueryEscape(entryURL))
	}
	query := parsed.Query()
	query.Set("url", entryURL)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
