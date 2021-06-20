package fromURI

/*
https://media.growth-and.com/go%E8%A8%80%E8%AA%9E%E3%81%A7rss%E3%83%95%E3%82%A3%E3%83%BC%E3%83%89%E3%82%92%E5%8F%96%E5%BE%97%E3%81%97%E3%81%A6%E3%81%BF%E3%82%8B/

存在確認 ファイル開く https://blog.lufia.org/entry/2019/05/07/234521
読み書き https://re-engines.com/2020/07/22/golang%E3%81%AE%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E6%93%8D%E4%BD%9C%E5%9F%BA%E6%9C%AC/
パイプの読み書きの方法もある https://waman.hatenablog.com/entry/2017/10/01/130330#osOpenFile-%E9%96%A2%E6%95%B0
全てまとまってる気がする

todo
 * マージする処理は『クロージャ』で書くとシンプルになりそう
 * 「フィードのファイルをざっくりパイプに投げる」と「パイプをバッファリングしながらフィードをまとめて一つのフィードにまとめる」を作りたい
  1. ストリームで渡ってきたテキスト( 連続したテキスト )を分離する処理
  2. 分離したフィードをsortableFeedに入れる、マージする
  3. それを標準出力で返す処理、あるいは構造体で返すべきだろうか? golang内であれば構造体で返した方がラク。
   * 最後にファイルなり標準出力に出力する関数は別に作る
 * 引数の処理
  * jqコマンドの不便さは | で区切れない( オプションを " " でくくる必要があるところかな、と思う )
  * パイプの処理を学ぶ
   * パイプで実行されたのか?を知る方法
  * https://orebibou.com/ja/home/201906/20190611_002/
  * https://kaneshin.hateblo.jp/entry/2016/07/05/004601
  * https://qiita.com/tanksuzuki/items/e712717675faf4efb07a
 * パッケージを分離する。マージとストアは分けたい。パイプやファイルからの読み込みも分離。
 * 「リモートのフィードが消えた場合」を実装する
 * フィードの時刻時刻がnowを修正したい。時刻をパースしてtimeの形式に変更する
 * デバックフラグをつけたい
 * 実行コマンドのファイル名はシンボリックリンク名を取得できるか? https://golang.hateblo.jp/entry/2018/10/22/080000
  * シンボリックリンクを沢山つくるようにしたい。busyboxみたく。で、名前は取得できるか。
 * 差分をSlack
 * s3への書き出し( httpで公開して、リーダーで読みたい )
  * golangでのawsライブラリ
 * オプションをyamlで設定できるように
 * はてブの自分のブックマークのRSSを監視。追加されたら、差分が発生したら、そのブックマークへのコメントをSlackへ
  * つまりローカルに「自分のブックマークのRSS」が保存されており、差分をマージ、更にそれらのブックマークを監視する
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

type sortableFeed struct {
	gofeed.Feed
}

func (b sortableFeed) Len() int {
	return len(b.Items)
}

func (b sortableFeed) Swap(i, j int) {
	b.Items[i], b.Items[j] = b.Items[j], b.Items[i]
}

func (b sortableFeed) Less(i, j int) bool {
	return b.Items[i].Published > b.Items[j].Published
}

/*
フィードを取得してローカルに保存する

todo
「リモートに無くなった」というケースも実装する必要がある
*/
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
