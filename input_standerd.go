package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mmcdole/gofeed"
)

/*
標準入力からフィードを取得して sortableFeed で返す

todo 標準入力から複数のフィードが与えられた特に分割する処理を追加
*/
func read() []sortableFeed {

	fp := gofeed.NewParser()
	b, _ := ioutil.ReadAll(os.Stdin)
	newFeed, err6 := fp.ParseString(string(b))
	if err6 != nil {
		fmt.Println(err6)
	}
	c1 := sortableFeed{*newFeed}
	m1 := []sortableFeed{c1}
	return m1
}
