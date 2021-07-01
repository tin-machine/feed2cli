package main

import (
	"fmt"
	"os"
  "bufio"

	"github.com/mmcdole/gofeed"
)

/*
標準入力からフィードを取得して sortableFeed で返す

todo 標準入力から複数のフィードが与えられた特に分割する処理を追加
 Scannerを使うとトークンで区切れる https://zenn.dev/hsaki/books/golang-io-package/viewer/bufio
 Scannerの区切り文字列は変更できる https://baubaubau.hatenablog.com/entry/2017/11/17/214635
  rss , atom などで区切り文字列は変更する必要がある。
*/
func read() []sortableFeed {
	fp := gofeed.NewParser()
  newFeed, err6 := fp.ParseURL("https://b.hatena.ne.jp/entrylist/general.rss")

	// b, _ := ioutil.ReadAll(os.Stdin)
	// newFeed, err6 := fp.ParseString(string(b))
  /* 標準出力を全部読み込んでgofeedのパーサーに渡す処理。だが、複数のrssを処理できるようにしたいのでコメントアウト
  */
  /* 
  os.Stdin が既にos.Openで帰ってくる構造体と同じだったら、Scannerに渡しやすい
  https://golang.org/pkg/os/#pkg-variables
  os.Stdin の返り値はFile、os.Openの返り値もFile

  https://baubaubau.hatenablog.com/entry/2017/11/17/214635#%E6%9B%B8%E3%81%84%E3%81%A6%E3%81%BF%E3%81%9F%E3%82%BD%E3%83%BC%E3%82%B9%E5%85%A8%E9%83%A8
  を参考にSplitFuncを設定していく
  */
  scanner := bufio.NewScanner(os.Stdin)
  scanner.Split(bufio.ScanWords)
  for scanner.Scan() {
    fmt.Println(scanner.Text())
    fmt.Println("区切ったよ")
  }

	if err6 != nil {
		fmt.Println(err6)
	}
	c1 := sortableFeed{*newFeed}
	m1 := []sortableFeed{c1}
	return m1
}
