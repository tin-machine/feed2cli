package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
	"golang.org/x/term"
)

func parseArgs() (isDebug, createSymlinks bool, operation, filterType string) {
	flag.BoolVar(&isDebug, "d", false, "Debug output")
	flag.BoolVar(&createSymlinks, "s", false, "Create symbolic links")
	flag.StringVar(&operation, "o", "", "Operation: merge, diff, slack, or hatena")
	flag.StringVar(&filterType, "f", "", "Filter to apply: hatena_bookmark")
	flag.Parse()
	return
}

func printDebugArgs() {
	for i, v := range os.Args {
		fmt.Printf("args[%d] -> %s\n", i, v)
	}
}

func createSymlinksIfNeeded() {
	_ = os.Symlink("feed2cli", "mergeRss")
	_ = os.Symlink("feed2cli", "diffRss")
	_ = os.Symlink("feed2cli", "slackRss")
	_ = os.Symlink("feed2cli", "hatenaRss")
}

// applyFilter は、フィードのリストを受け取り、指定されたフィルタを適用して
// FilteredItemのリストを返します。
func applyFilter(filterType string, feeds []*gofeed.Feed) ([]*FilteredItem, error) {
	var f Filter
	switch filterType {
	case "hatena_bookmark":
		f = &HatenaBookmarkFilter{}
	default:
		// 未対応のフィルタの場合は何もしないが、型を変換する必要がある
		items := []*gofeed.Item{}
		for _, feed := range feeds {
			items = append(items, feed.Items...)
		}
		filteredItems := make([]*FilteredItem, len(items))
		for i, item := range items {
			filteredItems[i] = &FilteredItem{Item: item}
		}
		return filteredItems, nil
	}

	// 全てのフィードからアイテムを一旦一つのスライスにまとめる
	allItems := []*gofeed.Item{}
	for _, feed := range feeds {
		allItems = append(allItems, feed.Items...)
	}

	return f.Apply(allItems)
}

// dispatchOperation は、操作とデータを受け取り、適切な出力関数にディスパッチします。
// データはフィルタリング済みか否かで型が異なるため、interface{}で受け取ります。
func dispatchOperation(operation, cmd string, data interface{}) {
	switch {
	case cmd == "mergeRss" || operation == "merge":
		// mergeとdiffはgofeed.Feedを期待するため、型変換が必要
		feeds := convertToFeeds(data)
		merged := Merge(feeds)
		OutputStanderd(merged)
	case cmd == "diffRss" || operation == "diff":
		feeds := convertToFeeds(data)
		diffed := Diff(feeds)
		OutputStanderd(diffed)
	case cmd == "slackRss" || operation == "slack":
		// slackはgofeed.Feedを期待する
		feeds := convertToFeeds(data)
		OutputSlack(feeds)
	case cmd == "hatenaRss" || operation == "hatena":
		// hatenaはFilteredItemを期待する
		if items, ok := data.([]*FilteredItem); ok {
			OutputHatenaToSlack(items)
		} else {
			fmt.Println("hatena操作にはフィルタリングされたデータが必要です。")
		}
	default:
		// デフォルトは標準出力
		OutputStanderd(data)
	}
}

func main() {
	isDebug, createSymlinks, operation, filterType := parseArgs()
	if isDebug {
		printDebugArgs()
	}
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
		if createSymlinks {
			createSymlinksIfNeeded()
		}
		if operation == "" && filterType == "" {
			fmt.Println("操作またはフィルタを指定してください: -o <operation> | -f <filter>")
			return
		}
	}

	if !term.IsTerminal(0) {
		s := read()
		cmd := strings.TrimLeft(os.Args[0], "./")

		if filterType != "" {
			filteredItems, err := applyFilter(filterType, s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "フィルタの適用に失敗しました: %v\n", err)
				os.Exit(1)
			}
			dispatchOperation(operation, cmd, filteredItems)
		} else {
			dispatchOperation(operation, cmd, s)
		}
	}
}

// convertToFeeds は、様々な型のデータを[]*gofeed.Feedに変換します。
// これは、mergeやdiffなど、gofeed.Feedを直接操作する既存の関数との互換性を保つためです。
func convertToFeeds(data interface{}) []*gofeed.Feed {
	if feeds, ok := data.([]*gofeed.Feed); ok {
		return feeds
	}

	if items, ok := data.([]*FilteredItem); ok {
		feedItems := make([]*gofeed.Item, len(items))
		for i, item := range items {
			feedItems[i] = item.Item
		}
		// 元のフィード情報は失われるが、一つのフィードにまとめる
		return []*gofeed.Feed{{Items: feedItems}}
	}
	
	if items, ok := data.([]*gofeed.Item); ok {
		return []*gofeed.Feed{{Items: items}}
	}

	return nil
}