package main

import (
	"fmt"

	"golang.org/x/term"
)

/*

todo
 * argv[0]で振る舞いを変えたい
  * 実行コマンドのファイル名はシンボリックリンク名を取得できるか?
   * シンボリックリンクを沢山つくるようにしたい。busyboxみたく。で、名前は取得できるか。
 * マージする処理は『クロージャ』で書くとシンプルになりそう
 * 「インプット」「操作(フィルター、マージ」「アウトプット」がわかりやすいかも。UNIX哲学を見てから決める
 * 引数の処理
  * jqコマンドの不便さは | で区切れない( オプションを " " でくくる必要があるところかな、と思う )
  * パイプの処理を学ぶ
 * 「リモートのフィードが消えた場合」を実装する
 * フィードの時刻時刻がnowを修正したい。時刻をパースしてtimeの形式に変更する
 * デバックフラグをつけたい
 * 差分をSlack
 * s3への書き出し( httpで公開して、リーダーで読みたい )
  * golangでのawsライブラリ
 * オプションをyamlで設定できるように
 * はてブの自分のブックマークのRSSを監視。追加されたら、差分が発生したら、そのブックマークへのコメントをSlackへ
  * つまりローカルに「自分のブックマークのRSS」が保存されており、差分をマージ、更にそれらのブックマークを監視する
*/

// main 関数は、コマンドラインからのインプットを受け取り、
// パイプが使用されているかどうかに応じて適切な処理を行います。
func main() {
	// パイプのある無しで振る舞いを変える
	// fmt.Println(terminal.IsTerminal(0))
	if term.IsTerminal(0) {
		fmt.Println("パイプ無し(FD値0)")
	} else {
		s := read()

		OutputSlack(s)
	}
	// StoreFeed("https://b.hatena.ne.jp/entrylist/general.rss", "feeds")
	// StoreFeed("https://b.hatena.ne.jp/entrylist/it.rss", "feeds")
}
