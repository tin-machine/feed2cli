package main

import (
	"log"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Diff は、二つのフィードを受け取り、古いフィードに存在するが新しいフィードには存在しないアイテムのリストを返します。
// 引数 fs は、古いフィードが fs[0] に、新しいフィードが fs[1] に格納されていることを前提としています。
// 戻り値として、新しいフィードに含まれないアイテムだけを持つ新しいフィードを返します。
func Diff(fs []*gofeed.Feed) []*gofeed.Feed {
	fp := gofeed.NewParser()

	// returnするフィードを作る
	feedData := `<feed xmlns="http://www.w3.org/2005/Atom"><subtitle>diff Atom</subtitle></feed>`

	diffFeed, err := fp.Parse(strings.NewReader(feedData))
	if err != nil {
		log.Fatalf("Feed parsing failed: %v", err) //  エラー処理を追加
	}

	// fs[0]とfs[1]のアイテムの差分を取る
	for _, oldItem := range fs[0].Items {
		existsInNewFeed := false
		for _, newItem := range fs[1].Items {
			if oldItem.Link == newItem.Link {
				existsInNewFeed = true
				break
			}
		}
		if !existsInNewFeed {
			diffFeed.Items = append(diffFeed.Items, oldItem)
		}
	}

	// sortableFeedに変換した後にソートする
	diffSortableFeed := sortableFeed{*diffFeed}
	diffSortableFeed.Sort() // フィードをソートする

	// 差分フィードを返す
	return []*gofeed.Feed{&diffSortableFeed.Feed}
}
