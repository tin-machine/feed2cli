package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/k0kubun/pp"
	"github.com/mmcdole/gofeed"
)

func Diff(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()

	// returnするフィードを作る
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom">
  <subtitle>diff Atom</subtitle>
  </feed>`

	diffFeed, _ := fp.Parse(strings.NewReader(feedData))

	/*
	  ここに差分を取る処理 Mergeの処理にかなり近い。同じURLか違うURLか
	  fs[0]、１つ目のフィードに古い方、fs[1]に新しいフィードが入っている前提
	*/
	for _, f0 := range fs[0].Items {
		addFlag := true
		// １つ目のフィードと差分を比較
		for _, f1 := range fs[1].Items {
			if f0.Link == f1.Link {
				addFlag = false
				break
			}
		}
		if addFlag {
			diffFeed.Items = append(diffFeed.Items, f0)
		}
	}

	sort.Sort(diffFeed)

	output_feed := []*gofeed.Feed{diffFeed}

	// デバック用
	if len(os.Args) > 1 && os.Args[1] == "-d" {
		fmt.Printf("diff.go で fs( 入力されたfeed)の個数は %d\n", len(fs))
		fmt.Printf("diff.go で fs[0].Items の個数は %d\n", len(fs[0].Items))
		fmt.Printf("diff.go で fs[1].Items の個数は %d\n", len(fs[1].Items))
		fmt.Printf("diff.go で output_feed の個数は %d\n", len(output_feed))
		fmt.Printf("diff.go で output_feed.Items の個数は %d\n", len(output_feed[0].Items))
		// pp.Print(output_feed)
	}

	return output_feed
}
