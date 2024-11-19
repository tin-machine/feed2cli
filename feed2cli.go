package main

import (
	"fmt"

	"golang.org/x/term"
)

// main 関数は、コマンドラインからのインプットを受け取り、
// パイプが使用されているかどうかに応じて適切な処理を行います。
func main() {
	// パイプのある無しで振る舞いを変える
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
	} else {
		s := read()

		OutputSlack(s)
	}
}
