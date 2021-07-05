package main

import (
	"fmt"

	"golang.org/x/crypto/ssh/terminal"
)

/*
https://media.growth-and.com/go%E8%A8%80%E8%AA%9E%E3%81%A7rss%E3%83%95%E3%82%A3%E3%83%BC%E3%83%89%E3%82%92%E5%8F%96%E5%BE%97%E3%81%97%E3%81%A6%E3%81%BF%E3%82%8B/

存在確認 ファイル開く https://blog.lufia.org/entry/2019/05/07/234521
読み書き https://re-engines.com/2020/07/22/golang%E3%81%AE%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E6%93%8D%E4%BD%9C%E5%9F%BA%E6%9C%AC/
パイプの読み書きの方法もある https://waman.hatenablog.com/entry/2017/10/01/130330#osOpenFile-%E9%96%A2%E6%95%B0
全てまとまってる気がする

todo
 * 標準入力から複数のフィードが与えられたら分割する
 * 関数を渡っていく変数はポインタの方が、効率が良さそう。
  * ほぼ同じ内容のメモリを関数に渡す( コピーする )よりはポインタだけが渡り、同じアドレス上のデータをどんどん上書きしていく。
  * ポインタを返すように関数を書き直す
 * マージする処理は『クロージャ』で書くとシンプルになりそう
 * 「フィードのファイルをざっくりパイプに投げる」と「パイプをバッファリングしながらフィードをまとめて一つのフィードにまとめる」を作りたい
  1. 分離したフィードをsortableFeedに入れる、マージする
  2. それを標準出力で返す処理、あるいは構造体で返すべきだろうか? golang内であれば構造体で返した方がラク。
   * 最後にファイルなり標準出力に出力する関数は別に作る
 * 「インプット」「操作(フィルター、マージ」「アウトプット」がわかりやすいかも。UNIX哲学を見てから決める
 * 引数の処理
  * jqコマンドの不便さは | で区切れない( オプションを " " でくくる必要があるところかな、と思う )
  * パイプの処理を学ぶ
   * パイプで実行されたのか?を知る方法
  * https://orebibou.com/ja/home/201906/20190611_002/
  * https://kaneshin.hateblo.jp/entry/2016/07/05/004601
  * https://qiita.com/tanksuzuki/items/e712717675faf4efb07a
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

func main() {
	// パイプのある無しで振る舞いを変える https://qiita.com/tanksuzuki/items/e712717675faf4efb07a
	fmt.Println(terminal.IsTerminal(0))
	if terminal.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
	} else {
		//b, _ := ioutil.ReadAll(os.Stdin)
		//fmt.Println("パイプで渡された内容(FD値0以外):", string(b))
		s := read()
		OutputStanderd(s)
	}
	StoreFeed("https://b.hatena.ne.jp/entrylist/general.rss", "feeds")
	StoreFeed("https://b.hatena.ne.jp/entrylist/it.rss", "feeds")
}
