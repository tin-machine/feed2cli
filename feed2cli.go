package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
  /*
  最初のオプションとして -d が与えられていたらデバック出力
  */
	if len(os.Args) >= 2 && os.Args[1] == "-d" {
		for i, v := range os.Args {
			fmt.Printf("args[%d] -> %s\n", i, v)
		}
	}

	// パイプのある無しで振る舞いを変える
	if terminal.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
	} else {
		//fmt.Println("パイプで渡された内容(FD値0以外):", string(b))
		s := read()
		switch os.Args[0] {
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
