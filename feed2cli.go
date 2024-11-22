package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// main 関数は、コマンドラインからのインプットを受け取り、
// パイプが使用されているかどうかに応じて適切な処理を行います。
func main() {
	// コマンドラインオプションの定義
	var isDebug bool
	var createSymlinks bool
	var operation string

	flag.BoolVar(&isDebug, "d", false, "Debug output")
	flag.BoolVar(&createSymlinks, "s", false, "Create symbolic links")
	flag.StringVar(&operation, "o", "", "Operation: merge, diff, or slack")
	flag.Parse()

	//  最初のオプションとして -d が与えられていたらデバック出力
	if isDebug {
		for i, v := range os.Args {
			fmt.Printf("args[%d] -> %s\n", i, v)
		}
	}
	// パイプのある無しで振る舞いを変える
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
		// -s だったらシンボリックリンクを作成する
		if createSymlinks {
			os.Symlink("feed2cli", "mergeRss")
			os.Symlink("feed2cli", "diffRss")
			os.Symlink("feed2cli", "slackRss")
		}
		// コマンドが指定されていない場合の処理
		if operation == "" {
			fmt.Println("操作を指定してください: merge, diff または slack")
			return
		}
	} else {
		// input_standerd.go にある read() を用いてフィードを分割
		s := read()
		// カレントディレクトリにシンボリックリンクを作ってある場合 ./ を削除
		cmd := strings.TrimLeft(os.Args[0], "./")

		if cmd == "mergeRss" || operation == "merge" {
			merged := Merge(s)
			OutputStanderd(merged)
		} else if cmd == "diffRss" || operation == "diff" {
			diffed := Diff(s)
			OutputStanderd(diffed)
		} else if cmd == "slackRss" || operation == "slack" {
			OutputSlack(s)
		} else {
			fmt.Println("無効なオプションです。使用可能なオプション: merge, diff, slack")
		}
	}
}
