package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/mmcdole/gofeed"
)

func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()

	// returnするフィードを作る
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom">
  <subtitle>Example Atom</subtitle>
  </feed>`

	mergedFeed, _ := fp.Parse(strings.NewReader(feedData))

	// 引数で与えられたいくつかのフィード
	for _, v := range fs {
		// フィードの中のアイテム
		for _, f := range v.Items {
			addFlag := true
			// マージ用のフィードのアイテムと比較していく
			for j, _ := range mergedFeed.Items {
				if mergedFeed.Items[j].Link == f.Link {
					// 同じURLのフィードは追加しない、次のフィードのチェックを行う
					addFlag = false
					break
				}
			}
			// 同じURLがなかった( trueのままだった )ら、そのフィードを追加
			if addFlag {
				mergedFeed.Items = append(mergedFeed.Items, f)
			}
		}
	}
	sort.Sort(mergedFeed)

	output_feed := []*gofeed.Feed{mergedFeed}
	// デバック用
	if len(os.Args) >= 2 && os.Args[1] == "-d" {
	}
	if len(os.Args) > 1 && os.Args[1] == "-d" {
		fmt.Printf("merge.go で fs( 入力されたfeed)の個数は %d\n", len(fs))
		fmt.Printf("merge.go で fs[0].Items の個数は %d\n", len(fs[0].Items))
		fmt.Printf("merge.go で fs[1].Items の個数は %d\n", len(fs[1].Items))
		fmt.Printf("merge.go で output_feed の個数は %d\n", len(output_feed))
		fmt.Printf("merge.go で output_feed.Items の個数は %d\n", len(output_feed[0].Items))
		// pp.Print(output_feed)
		fmt.Printf("Merge で []*gofeed.Feed の個数は %d\n[]*gofeed.Feed の中身は ↓", len(fs))
		pp.Print(fs)
	}

	return output_feed
}
