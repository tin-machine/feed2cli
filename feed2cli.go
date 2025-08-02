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
	flag.StringVar(&operation, "o", "", "Operation: merge, diff, slack, hatena, json")
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

func applyFilter(filterType string, feeds []*gofeed.Feed) ([]*FilteredItem, error) {
	var f Filter
	switch filterType {
	case "hatena_bookmark":
		f = &HatenaBookmarkFilter{}
	default:
		// フィルタが指定されていない、または未対応の場合は、型変換のみ行う
		return convertToFilteredItems(feeds), nil
	}

	allItems := []*gofeed.Item{}
	for _, feed := range feeds {
		allItems = append(allItems, feed.Items...)
	}

	return f.Apply(allItems)
}

func dispatchOperation(operation, cmd string, data interface{}) {
	op := operation
	if op == "" {
		op = strings.TrimSuffix(cmd, "Rss")
	}

	switch op {
	case "merge":
		feeds := convertToFeeds(data)
		merged := Merge(feeds)
		OutputStanderd(merged)
	case "diff":
		feeds := convertToFeeds(data)
		diffed := Diff(feeds)
		OutputStanderd(diffed)
	case "slack":
		OutputSlack(data)
	case "hatena":
		if items, ok := data.([]*FilteredItem); ok {
			OutputHatenaToSlack(items)
		} else {
			fmt.Fprintln(os.Stderr, "hatena操作にはフィルタリングされたデータが必要です。-f hatena_bookmark を使用してください。")
		}
	case "json":
		OutputJSON(data)
	default:
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

		var dataToDispatch interface{}
		dataToDispatch = s // デフォルトは元のフィード

		if filterType != "" {
			filteredItems, err := applyFilter(filterType, s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "フィルタの適用に失敗しました: %v\n", err)
				os.Exit(1)
			}
			dataToDispatch = filteredItems
		}

		dispatchOperation(operation, cmd, dataToDispatch)
	}
}

func convertToFeeds(data interface{}) []*gofeed.Feed {
	if feeds, ok := data.([]*gofeed.Feed); ok {
		return feeds
	}

	if items, ok := data.([]*FilteredItem); ok {
		feedItems := make([]*gofeed.Item, len(items))
		for i, item := range items {
			feedItems[i] = item.Item
		}
		return []*gofeed.Feed{{Items: feedItems}}
	}

	if items, ok := data.([]*gofeed.Item); ok {
		return []*gofeed.Feed{{Items: items}}
	}

	return nil
}
