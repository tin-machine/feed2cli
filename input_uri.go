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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

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

	f := openOrCreateFileWithDirs(file, dir, 0)

	defer f.Close() // ファイルポインタは必ず閉じる

	// 存在する場合、古いフィードを読み込む
	oldFeed, err := fp.Parse(f)
	if err == nil {
		c2 := &sortableFeed{*oldFeed}
		for _, oldItem := range c2.Items {
			if !itemExists(c1.Items, oldItem.Link) {
				c1.Items = append(c1.Items, oldItem)
			}
		}
	}

	// マージ後にソートを確実に実行
	c1.Sort()

	// gofeed.Feed 構造体を直接JSONとして保存
	jsonData, err := json.MarshalIndent(c1.Feed, "", "  ")
	if err != nil {
		log.Fatalf("フィードのJSONエンコードに失敗しました: %v", err)
	}

	if err := os.WriteFile(file, jsonData, filePermission); err != nil {
		log.Fatalf("ファイルの書き込みに失敗しました: %v", err)
	}
}
