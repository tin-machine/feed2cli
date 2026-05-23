package main

import (
	"fmt"
	"os"

	"github.com/mmcdole/gofeed"
)

func itemExists(items []*gofeed.Item, link string) bool {
	target := normalizeFeedURL(link)
	for _, item := range items {
		if itemDedupKey(item) == target {
			return true
		}
	}
	return false
}

func openOrCreateFileWithDirs(filePath, dir string, perm os.FileMode) (*os.File, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("ディレクトリ作成に失敗しました: %w", err)
	}
	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, perm)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開く際にエラーが発生しました: %w", err)
	}
	return f, nil
}
