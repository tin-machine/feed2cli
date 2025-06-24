package main

import (
	"os"
	"log"
	"github.com/mmcdole/gofeed"
)

func itemExists(items []*gofeed.Item, link string) bool {
	for _, item := range items {
		if item.Link == link {
			return true
		}
	}
	return false
}

func openOrCreateFileWithDirs(filePath, dir string, perm os.FileMode) *os.File {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("ディレクトリ作成に失敗しました: %v", err)
	}
	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, perm)
	if err != nil {
		log.Fatalf("ファイルを開く際にエラーが発生しました: %v", err)
	}
	return f
}
