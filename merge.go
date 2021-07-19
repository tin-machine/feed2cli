package main

import (
  "strings"
  "sort"

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
  return output_feed
}
