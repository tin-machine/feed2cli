package main

import (
	"fmt"
	"os"
  "bufio"
  "strings"

	"github.com/mmcdole/gofeed"
)

func splitFeed(data []byte, atEOF bool) (advance int, token []byte, err error) {
  	// Return nothing if at end of file and no data passed
  	if atEOF && len(data) == 0 {
  		return 0, nil, nil
  	}
  	if i := strings.Index(string(data), "</rss>"); i >= 0 {
      end := i + len("</rss>")
  		return end, data[0:end], nil
  	}

  	if i := strings.Index(string(data), "</feed>"); i >= 0 {
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
標準入力からフィードを取得して スライスの中にsortableFeedのポインタが入った形で返す
*/
func read() []*sortableFeed {
	fp := gofeed.NewParser()

  // バッファサイズを大きくする必要があった https://mickey24.hatenablog.com/entry/bufio_scanner_line_length
  const (
    initialBufSize = 10000
    maxBufSize = 1000000
  )

  scanner := bufio.NewScanner(os.Stdin)
  buf := make([]byte, initialBufSize)
  scanner.Buffer(buf, maxBufSize)

  // 最後にreturnするためのスライス
  slice := []*sortableFeed{}
  scanner.Split(splitFeed)
  for scanner.Scan() {
    fmt.Print("\n\n\n区切ったよ\n\n\n")
    feed, _ := fp.ParseString(scanner.Text())
    c2 := &sortableFeed{*feed}
    slice = append(slice, c2)
  }

	return slice
}
