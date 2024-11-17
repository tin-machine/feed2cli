package main

import (
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Diff は、二つのフィードを受け取り、古いフィードに存在するが新しいフィードには存在しないアイテムのリストを返します。
// 引数 fs は、古いフィードが fs[0] に、新しいフィードが fs[1] に格納されていることを前提としています。
// 戻り値として、新しいフィードに含まれないアイテムだけを持つ新しいフィードを返します。
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
	return output_feed
}
