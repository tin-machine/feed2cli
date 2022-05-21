package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	_ "github.com/k0kubun/pp"
	"github.com/mmcdole/gofeed"
)

/*
標準入力から複数のfeedを連続して受け取れるように、
それぞれのfeedを分離できるようにする。
末尾が</rss>や</feed>で終わることを期待している

scanner.Split(splitFeed) という形で利用する。
https://qiita.com/k_dutch/items/4e983597ee091e517659
*/
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

	if i := strings.Index(string(data), "</rdf:RDF>"); i >= 0 {
		end := i + len("</rdf:RDF>")
		return end, data[0:end], nil
	}

	if atEOF {
		return len(data), data, nil
	}
	return
}

/*
絵文字が含まれるとgofeedがエラーになるので除去する
https://qiita.com/sshon/items/1f3b14aed47217c72242
*/
func printOnly(r rune) rune {
	if unicode.IsPrint(r) {
		return r
	}
	return -1
}

/*
標準入力からフィードを取得して gofeed.Feedのスライスで返す
*/
func read() []*gofeed.Feed {
	fp := gofeed.NewParser()

	// バッファサイズを大きくする必要があった https://mickey24.hatenablog.com/entry/bufio_scanner_line_length
	const (
		initialBufSize = 10000
		maxBufSize     = 1000000
	)

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, initialBufSize)
	scanner.Buffer(buf, maxBufSize)

	// 最後にreturnするためのスライス
	slice := []*gofeed.Feed{}
	scanner.Split(splitFeed)
	for scanner.Scan() {
		xmlData := strings.Map(printOnly, string(scanner.Text()))
		feed, err := fp.ParseString(xmlData)
		if err != nil {
			fmt.Fprintln(os.Stderr, os.Stderr)
		} else {
			slice = append(slice, feed)
		}
	}

	// デバック用コード
	if len(os.Args) > 1 && os.Args[1] == "-d" {
		fmt.Printf("input_standerd で slice の個数は %d\n", len(slice))
	}

	return slice
}
