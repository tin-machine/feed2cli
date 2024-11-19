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
	"log"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

// 定数を定義
const (
	filePermission = 0666 // ファイルのパーミッション
)

// StoreFeed は、指定された URL からフィードを取得し、ローカルに保存します。
// 既存のフィードと比較し、差分があれば更新を行います。
func StoreFeed(url string, prefix string) {
	fp := gofeed.NewParser()
	newFeed, err := fp.ParseURL(url)
	if err != nil {
		log.Fatalf("URLからフィードの取得に失敗しました: %v", err)
	}
	c1 := &sortableFeed{*newFeed}
	file := prefix + "/" + regexp.MustCompile(`http(s*):\/\/`).ReplaceAllString(c1.Link, "")
	dir := regexp.MustCompile(`(.*)/`).FindString(file)
	fmt.Println("ファイル名: ", file)
	fmt.Println("ディレクトリ名: ", dir)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil { // 追加: ディレクトリ作成時のエラーハンドリング
		log.Fatalf("ディレクトリ作成に失敗しました: %v", err)
	}

	// 既存のファイルがあるか確認
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("ファイルが存在しないので、新規作成します: %s\n", file)
	} else if err != nil {
		log.Fatalf("ファイルの状態確認時にエラー: %v", err)
	}

	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("ファイルが存在しないので、ディレクトリ作成します")
			err := os.MkdirAll(dir, os.ModePerm) // パーミッションを定数から取得
			if err != nil {
				log.Fatalf("ディレクトリ作成に失敗しました: %v", err)
			}
		} else {
			log.Fatalf("ファイルを開く際にエラーが発生しました: %v", err)
		}
	}

	defer f.Close() // ファイルポインタは必ず閉じる

	// 存在する場合、古いフィードを読み込む
	oldFeed, err := fp.Parse(f)
	if err == nil {
		c2 := &sortableFeed{*oldFeed}
		for _, oldItem := range c2.Items {
			addFlag := true
			for _, newItem := range c1.Items {
				if newItem.Link == oldItem.Link {
					addFlag = false
					break
				}
			}
			if addFlag {
				c1.Items = append(c1.Items, oldItem)
			}
		}
	}

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
		log.Fatalf("RSSの生成に失敗しました: %v", err)
	}

	if err := os.WriteFile(file, []byte(rss), filePermission); err != nil {
		log.Fatalf("ファイルの書き込みに失敗しました: %v", err)
	}
}
