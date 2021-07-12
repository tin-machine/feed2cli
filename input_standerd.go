package main

import (
	"fmt"
	"os"
  "bufio"
  "strings"

	"github.com/mmcdole/gofeed"
)

// https://stackoverflow.com/questions/33068644/how-a-scanner-can-be-implemented-with-a-custom-split/33069759
func SplitAt(substr string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
  return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
  	// Return nothing if at end of file and no data passed
  	if atEOF && len(data) == 0 {
  		return 0, nil, nil
  	}
  	// Find the index of the input of the separator substring
    // strings.Index()で文字列の位置を取得できる https://itsakura.com/go-index
    // この関数内で「RSS、Atomの末尾を処理する」
  	if i := strings.Index(string(data), substr); i >= 0 {
      // 区切り文字も含めたいので少し修正
      end := i + len(substr)
  		return end, data[0:end], nil
  	}
  	// If at end of file with data return the data
  	if atEOF {
  		return len(data), data, nil
  	}
  	return
  }
}

func splitFeed(data []byte, atEOF bool) (advance int, token []byte, err error) {
  	// Return nothing if at end of file and no data passed
  	if atEOF && len(data) == 0 {
  		return 0, nil, nil
  	}
    // strings.Index()で文字列の位置を取得できる https://itsakura.com/go-index
    // この関数内で「RSS、Atomの末尾を処理する」
  	if i := strings.Index(string(data), "</rss>"); i >= 0 {
      // 区切り文字も含めたいので少し修正
      end := i + len("</rss>")
  		return end, data[0:end], nil
  	}

  	if i := strings.Index(string(data), "</feed>"); i >= 0 {
      // 区切り文字も含めたいので少し修正
      end := i + len("</feed>")
  		return end, data[0:end], nil
  	}

  	// If at end of file with data return the data
  	if atEOF {
  		return len(data), data, nil
  	}
  	return
}

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

  // バッファサイズを大きくする必要があった https://mickey24.hatenablog.com/entry/bufio_scanner_line_length
  const (
    initialBufSize = 10000
    maxBufSize = 1000000
  )

  scanner := bufio.NewScanner(os.Stdin)
  buf := make([]byte, initialBufSize)
  scanner.Buffer(buf, maxBufSize)
  // 任意の文字列で区切りたい 
  // https://baubaubau.hatenablog.com/entry/2017/11/17/214635#%E6%9B%B8%E3%81%84%E3%81%A6%E3%81%BF%E3%81%9F%E3%82%BD%E3%83%BC%E3%82%B9%E5%85%A8%E9%83%A8
  // 任意の文字列を引数に与えると、splitFunctionな関数を返す関数
  // https://stackoverflow.com/questions/33068644/how-a-scanner-can-be-implemented-with-a-custom-split/33069759

  // scanner.Split(SplitAt("</rss>"))
  scanner.Split(splitFeed)
  // scanner.Split(bufio.ScanWords)
  for scanner.Scan() {
    fmt.Println(scanner.Text())
    fmt.Print("\n\n\n区切ったよ\n\n\n")
  }

	if err6 != nil {
		fmt.Println(err6)
	}
	c1 := sortableFeed{*newFeed}
	m1 := []sortableFeed{c1}
	return m1
}
