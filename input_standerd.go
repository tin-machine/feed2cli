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
  delim := []byte("</rss>")
  var splitFunction = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
    for i := 0; i < len(data); i++ {
      if data[i] == delim[0] && data[i+1] == delim[1] && data[i+2] == delim[2] && data[i+3] == delim[3] && data[i+4] == delim[4] && data[i+5] == delim[5] && data[i+6] == delim[6] && data[i+7] == delim[7] && data[i+8] == delim[8]  && data[i+9] == delim[9]  && data[i+10] == delim[10]  && data[i+11] == delim[11]  && data[i+12] == delim[12]  {
        return i + 12, data[:i+12], nil //tokenをdata[:i+3]としているので、区切り文字は含まれる
      }
    }
    return 0, data, bufio.ErrFinalToken
  }
  scanner.Split(splitFunction)
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
