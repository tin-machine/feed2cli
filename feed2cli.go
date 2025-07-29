package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	"github.com/mmcdole/gofeed"
)

// main 関数は、コマンドラインからのインプットを受け取り、
// パイプが使用されているかどうかに応じて適切な処理を行います。
func parseArgs() (isDebug, createSymlinks bool, operation string) {
	flag.BoolVar(&isDebug, "d", false, "Debug output")
	flag.BoolVar(&createSymlinks, "s", false, "Create symbolic links")
	flag.StringVar(&operation, "o", "", "Operation: merge, diff, slack, or hatena")
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

func dispatchOperation(operation, cmd string, s []*gofeed.Feed) {
	switch {
	case cmd == "mergeRss" || operation == "merge":
		merged := Merge(s)
		OutputStanderd(merged)
	case cmd == "diffRss" || operation == "diff":
		diffed := Diff(s)
		OutputStanderd(diffed)
	case cmd == "slackRss" || operation == "slack":
		OutputSlack(s)
	case cmd == "hatenaRss" || operation == "hatena":
		OutputHatenaToSlack(s)
	default:
		fmt.Println("無効なオプションです。使用可能なオプション: merge, diff, slack, hatena")
	}
}

func main() {
	isDebug, createSymlinks, operation := parseArgs()
	if isDebug {
		printDebugArgs()
	}
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
		if createSymlinks {
			createSymlinksIfNeeded()
		}
		if operation == "" {
			fmt.Println("操作を指定してください: merge, diff, slack, or hatena")
			return
		}
	} else {
		s := read()
		cmd := strings.TrimLeft(os.Args[0], "./")
		dispatchOperation(operation, cmd, s)
	}
}

