package main

/*
「URLから取得する」だけにする。
ローカルに貯める、は、
1. URLから取得する <= curl で良いかも
2. ローカルファイルから取得する <= cat で良いかも
3. マージする
4. マージしたデータをアウトプットする
この処理の組み合わせにする。

[1]と[2]はcurl, cat でできるが、一つコマンドで feedMerge や feedStore みたいなものがあると便利かも、そういう場合つくる
( かも、ではなくて、実際の需要がでるまで 標準入出力で進めた方がシンプルかも

貯めるだけ貯めて、teeコマンドと併用して次の標準出力に出せるようにすると良さそう

URLから取得する、は、curlで良いか、パイプで標準入力から受け取り、パースする処理から始める。

*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/feeds"

	"github.com/mmcdole/gofeed"
)

// StoreFeed は、指定された URL からフィードを取得し、ローカルに保存します。
// 既存のフィードと比較し、差分があれば更新を行います。
func StoreFeed(url string, prefix string) {
	/*
		リモートとローカルにフィードが存在する。
		ローカルにはフィードが存在しない可能性があるので、リモートから検出する
	*/

	fp := gofeed.NewParser()
	newFeed, err6 := fp.ParseURL(url)
	if err6 != nil {
		fmt.Println(err6)
	}
	c1 := &sortableFeed{*newFeed}

	file := prefix + "/" + regexp.MustCompile(`http(s*):\/\/`).ReplaceAllString(c1.Link, "")
	dir := regexp.MustCompile(`(.*)/`).FindString(file)
	fmt.Println(file)
	fmt.Println(dir)

	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("ファイルが開けませんでした")
		if os.IsNotExist(err) {
			// ファイルが存在しないので新しく作る
			// 先にディレクトリを作る
			// todo
			// 「ファイルは存在しない」が「ディレクトリは存在する」というケースに対応する
			fmt.Println("ファイルが存在しないので、ディレクトリ作成します")
			os.MkdirAll(dir, 0777)
		}
	} else {
		// ファイルが存在するので読み込み
		fmt.Println("ファイルが存在します")
		oldFeed, _ := fp.Parse(f)
		c2 := &sortableFeed{*oldFeed}
		for c2_i, _ := range c2.Items {
			addFlag := true
			for c1_i, _ := range c1.Items {
				// 下記は『同じURLが合ったらbreak、無かったら最後に追加する』という処理にする
				if c1.Items[c1_i].Link == c2.Items[c2_i].Link {
					addFlag = false
					break
				}
			}
			if addFlag {
				c1.Items = append(c1.Items, c2.Items[c2_i])
			}
		}
	}
	defer f.Close()

	now := time.Now()
	output_feed := &feeds.Feed{
		Title:       c1.Title,
		Link:        &feeds.Link{Href: c1.Link},
		Description: c1.Description,
		Created:     now,
	}

	for _, v := range c1.Items {
		item := &feeds.Item{
			Title:       v.Title,
			Link:        &feeds.Link{Href: v.Link},
			Description: v.Description,
			Created:     now,
		}
		output_feed.Add(item)
	}

	// RSS フォーマットに変換して保存
	rss, err := output_feed.ToRss()
	if err != nil {
		log.Fatal(err)
	}
	// string 型 → []byte 型
	b := []byte(rss)

	err2 := ioutil.WriteFile(file, b, 0666)
	if err2 != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}
