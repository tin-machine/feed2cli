# 概要

最近、RSS復活の兆しが少し見えます。
[事実上の「Googleリーダー」復活？ ChromeにRSSを用いたフォロー機能が搭載される](https://internet.watch.impress.co.jp/docs/yajiuma/1357390.html)

RSSは扱いやすいフォーマットで便利なのですが、
例えば、はてなブックマークなどを複数キーワードで追いかけようとすると、
別のRSSに同じエントリが入ってしまい無駄です。

手軽にRSSをマージできるツールが無いかと探しているのですが見つからなかったので
Golangの勉強がてらRSSのエントリをマージするツールを作成中です。

# ビルドとシンボリックリンク作成

```
$ make
```

このプログラムはBusyboxのように実行されたシンボリックリンクの名前で振る舞いを変えます。
最初にシンボリックリンクを作成するため -s オプションでシンボリックリンクを作成してください。

```
$ ./feed2cli -s
```

結果

# 使い方

はてなブックマークでは気になるキーワードの検索結果
[例えばlinux](https://b.hatena.ne.jp/search/text?q=linux&users=500)に
『&mode=rss』を付けるだけでRSSになります。

例)
* linux https://b.hatena.ne.jp/search/text?q=linux&users=500&mode=rss'
* docker https://b.hatena.ne.jp/search/text?q=docker&users=500&mode=rss'

↑　上記を ↓ のようにマージします。

```
$ echo $(curl -L 'http://b.hatena.ne.jp/hotentry/it.rss') $(curl -L 'http://b.hatena.ne.jp/hotentry.rss') | ./mergeRss > test.rss
```

