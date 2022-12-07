### 概要

最近、RSS復活の兆しが少し見えます。

[事実上の「Googleリーダー」復活？ ChromeにRSSを用いたフォロー機能が搭載される](https://internet.watch.impress.co.jp/docs/yajiuma/1357390.html)

ただ... はてなブックマークなどを複数キーワードで追いかけようとすると、
別のRSSに同じエントリが入ってしまい無駄です。

手軽にRSSをマージできるツールが無いかと探しているのですが見つからなかったので
Golangの勉強がてらRSSのエントリをマージするツールを作成中です。

### ビルドとシンボリックリンク作成

```
$ make
```

このプログラムはBusyboxのように実行されたシンボリックリンクの名前で振る舞いを変えます。
最初にシンボリックリンクを作成するため -s オプションでシンボリックリンクを作成してください。

```
$ ./feed2cli -s
```

結果

### 使い方

はてなブックマークでは気になるキーワードの検索結果、[例えばlinux](https://b.hatena.ne.jp/search/text?q=linux&users=500)に

『&mode=rss』を付けるだけでRSSになります。 [参考](https://anond.hatelabo.jp/20220521220951)

例)
* linux https://b.hatena.ne.jp/search/text?q=linux&users=500&mode=rss'
* docker https://b.hatena.ne.jp/search/text?q=docker&users=500&mode=rss'

↑　上記を ↓ のようにマージします。

```
$ echo $(curl -L 'http://b.hatena.ne.jp/hotentry/it.rss') \
  $(curl -L 'http://b.hatena.ne.jp/hotentry.rss') \
  | ./mergeRss > test.rss
```

### 比較方法

単純にURLが同じだったらマージしています。

### どの位、減ってるの?

デバック用オプション -d を付けると、各RSSにどの程度エントリが含まれているかを表示しています。
ですので ↓ のようにすると

```
$ echo \
  $(curl 'https://b.hatena.ne.jp/search/text?q=linux&users=500&mode=rss') \
  $(curl 'https://b.hatena.ne.jp/search/text?q=docker&users=500&mode=rss') \
  $(curl 'https://b.hatena.ne.jp/search/text?q=docker-compose&users=500&mode=rss') \
  | ./mergeRss -d | less
```

下記の出力がでてきます。検索結果のページは40個のエントリが含まれていて、
合計120個になるところ、98個まで減りました( 2022/05/22時点 )

```
input_standerd で slice の個数は 3
Merge で fs( 入力されたfeed)の個数は 3
Merge で 何番目のfeedを処理しているか? 0
Merge で fs[0].Items の個数は 40
Merge で 何番目のfeedを処理しているか? 1
Merge で fs[1].Items の個数は 40
Merge で 何番目のfeedを処理しているか? 2
Merge で fs[2].Items の個数は 40
Merge で output_feed の個数は 1
Merge で output_feed.Items の個数は 98
Merge で []*gofeed.Feed の個数は 3
```
