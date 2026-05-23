package main

/*
「URLから取得する」だけにする。
ローカルに貯める、は、
1. URLから取得する <= curl で良いかも
2. ローカルファイルから取得する <= cat で良いかも
3. マージする
4. マージしたデータをアウトプットする
この処理の組み合わせにする。

[1]と[2]はcurl, cat でできるが、一つコマンドで feedMerge や feedStore みたいなものがあると便利かも、そういう場合つくる
( かも、ではなくて、実際の需要がでるまで 標準入出力で進めた方がシンプルかも

貯めるだけ貯めて、teeコマンドと併用して次の標準出力に出せるようにすると良さそう

URLから取得する、は、curlで良いか、パイプで標準入力から受け取り、パースする処理から始める。

*/

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/mmcdole/gofeed"
)

// 定数を定義
const (
	filePermission = 0666 // ファイルのパーミッション
)

type StoreFeedOptions struct {
	Prefix string
	Fetch  func(string) (*gofeed.Feed, error)
}

// StoreFeed は、指定された URL からフィードを取得し、ローカルに保存します。
// 既存のフィードと比較し、差分があれば更新を行います。
func StoreFeed(url string, prefix string) {
	if _, err := StoreFeedWithOptions(url, StoreFeedOptions{Prefix: prefix}); err != nil {
		log.Fatal(err)
	}
}

func StoreFeedWithOptions(feedURL string, options StoreFeedOptions) (string, error) {
	fetch := options.Fetch
	if fetch == nil {
		fetch = fetchFeedFromURL
	}

	newFeed, err := fetch(feedURL)
	if err != nil {
		return "", fmt.Errorf("URLからフィードの取得に失敗しました: %w", err)
	}
	if newFeed == nil {
		return "", fmt.Errorf("URLからフィードの取得に失敗しました: empty feed")
	}

	file, dir := storeFeedPath(options.Prefix, newFeed)
	fmt.Println("ファイル名: ", file)
	fmt.Println("ディレクトリ名: ", dir)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("ディレクトリ作成に失敗しました: %w", err)
	}

	oldFeed, err := loadStoredFeed(file)
	if err != nil {
		return "", err
	}

	c1 := &sortableFeed{*newFeed}
	if oldFeed != nil {
		c2 := &sortableFeed{*oldFeed}
		for _, oldItem := range c2.Items {
			if !itemExists(c1.Items, oldItem.Link) {
				c1.Items = append(c1.Items, oldItem)
			}
		}
	}

	// マージ後にソートを確実に実行
	c1.Sort()

	// gofeed.Feed 構造体を直接JSONとして保存
	jsonData, err := json.MarshalIndent(c1.Feed, "", "  ")
	if err != nil {
		return "", fmt.Errorf("フィードのJSONエンコードに失敗しました: %w", err)
	}

	if err := os.WriteFile(file, jsonData, filePermission); err != nil {
		return "", fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}
	return file, nil
}

func fetchFeedFromURL(feedURL string) (*gofeed.Feed, error) {
	return gofeed.NewParser().ParseURL(feedURL)
}

func fetchFeedsFromURLs(feedURLs []string) ([]*gofeed.Feed, error) {
	feeds := make([]*gofeed.Feed, 0, len(feedURLs))
	for _, feedURL := range feedURLs {
		feed, err := fetchFeedFromURL(feedURL)
		if err != nil {
			return nil, fmt.Errorf("URLからフィードの取得に失敗しました (%s): %w", feedURL, err)
		}
		if feed == nil {
			return nil, fmt.Errorf("URLからフィードの取得に失敗しました (%s): empty feed", feedURL)
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func storeFeedPath(prefix string, feed *gofeed.Feed) (file, dir string) {
	c1 := &sortableFeed{*feed}
	file = prefix + "/" + regexp.MustCompile(`http(s*):\/\/`).ReplaceAllString(c1.Link, "")
	dir = regexp.MustCompile(`(.*)/`).FindString(file)
	return file, dir
}

func loadStoredFeed(file string) (*gofeed.Feed, error) {
	f, err := openOrCreateFileWithDirs(file, regexp.MustCompile(`(.*)/`).FindString(file), filePermission)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("ファイルの状態確認時にエラー: %w", err)
	}
	if stat.Size() == 0 {
		fmt.Printf("ファイルが存在しないので、新規作成します: %s\n", file)
		return nil, nil
	}

	var stored gofeed.Feed
	if err := json.NewDecoder(f).Decode(&stored); err != nil {
		return nil, fmt.Errorf("保存済みフィードJSONの読み込みに失敗しました: %w", err)
	}
	return &stored, nil
}
