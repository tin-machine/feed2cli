package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/mmcdole/gofeed"
)

// splitFeed は、標準入力からのデータをパースして、<rss> または <feed> タグで終わる部分を抽出します。
// atEOF が true の場合、データの末尾であると考え、非表示のデータも返します。
func splitFeed(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
	if atEOF {
		return len(data), data, nil
	}
	return
}

// printOnly は、印刷可能な文字のみを返し、絵文字などの非印刷可能な文字を除外します。
func printOnly(r rune) rune {
	if unicode.IsPrint(r) {
		return r
	}
	return -1
}

// read は、標準入力から取得したデータを gofeed.Feed のスライスとして返します。
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

	// スライスの初期化
	slice := []*gofeed.Feed{}
	scanner.Split(splitFeed)

	// スキャナーでフィードを読込んで、パースする
	for scanner.Scan() {
		xmlData := strings.Map(printOnly, string(scanner.Text()))
		feed, err := fp.ParseString(xmlData)
		if err != nil {
			log.Printf("フィードのパースに失敗しました: %v", err) // エラーハンドリングを追加
			continue                               // 無効なフィードがあっても次に進む
		}
		slice = append(slice, feed)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("スキャナーの読み込み時にエラーが発生しました: %v", err)
	}

	return slice
}
