package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/k0kubun/pp"
	"github.com/mmcdole/gofeed"
)

func Merge(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()
	// デバック用
	if len(os.Args) >= 2 && os.Args[1] == "-d" {
		fmt.Printf("Merge で fs( 入力されたfeed)の個数は %d\n", len(fs))
	}

	// returnするフィードを作る
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom">
  <subtitle>Example Atom</subtitle>
  </feed>`

	mergedFeed, _ := fp.Parse(strings.NewReader(feedData))

	// 引数で与えられたいくつかのフィード
	for i, v := range fs {
		// デバック用コード
		if len(os.Args) > 1 && os.Args[1] == "-d" {
			// フィードの中のアイテム
			fmt.Printf("Merge で 何番目のfeedを処理しているか? %d\n", i)
			fmt.Printf("Merge で fs[%d].Items の個数は %d\n", i, len(fs[0].Items))
		}
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
	if len(os.Args) > 1 && os.Args[1] == "-d" {
		fmt.Printf("Merge で output_feed の個数は %d\n", len(output_feed))
		fmt.Printf("Merge で output_feed.Items の個数は %d\n", len(output_feed[0].Items))
		fmt.Printf("Merge で []*gofeed.Feed の個数は %d\n[]*gofeed.Feed の中身は ↓", len(fs))
	}

	return output_feed
}
