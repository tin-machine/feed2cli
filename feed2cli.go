package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// main 関数は、コマンドラインからのインプットを受け取り、
// パイプが使用されているかどうかに応じて適切な処理を行います。
func main() {
	//  最初のオプションとして -d が与えられていたらデバック出力
	if len(os.Args) >= 2 && os.Args[1] == "-d" {
		for i, v := range os.Args {
			fmt.Printf("args[%d] -> %s\n", i, v)
		}
	}
	// パイプのある無しで振る舞いを変える
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
		// -s だったらシンボリックリンクを作成する
		if len(os.Args) > 1 && os.Args[1] == "-s" {
			os.Symlink("feed2cli", "mergeRss")
			os.Symlink("feed2cli", "diffRss")
			os.Symlink("feed2cli", "slackRss")
		}
	} else {
		// input_standerd.go にある read() を用いてフィードを分割
		s := read()
		// カレントディレクトリにシンボリックリンクを作ってある場合 ./ を削除
		cmd := strings.TrimLeft(os.Args[0], "./")
		switch cmd {
		case "mergeRss":
			merged := Merge(s)
			OutputStanderd(merged)
		case "diffRss":
			diffed := Diff(s)
			OutputStanderd(diffed)
		case "slackRss":
			OutputSlack(s)
		}
	}
}
