# feed2cli

### 概要

RSS/Atomフィードをマージ、差分、フィルタリングするためのCLIツールです。
Plaggerにインスパイアされており、Unixのパイプライン思想に沿って、手軽にフィードを加工することを目指しています。

例えば、はてなブックマークの複数キーワードの検索結果を追いかける際に、重複するエントリをマージしたり、各エントリにはてなブックマークのコメントを付与したり、その結果をSlackに通知したりできます。

### 主な機能

*   **merge**: 複数のフィードをマージし、重複エントリを排除します。
*   **diff**: 2つのフィードを比較し、新しいフィードにのみ存在するエントリを抽出します。
*   **filter**: `-f`フラグで指定されたフィルタを適用し、フィードの各エントリに追加情報を付与します。
    *   `hatena_bookmark`: はてなブックマークのブックマーク数とコメントを取得します。
*   **output**: `-o`フラグで指定された形式で結果を出力します。
    *   `(デフォルト)`: RSS形式で標準出力します。
    *   `slack`: Slackに汎用的なメッセージを通知します。
    *   `hatena`: はてなブックマークのコメントをSlackのスレッドに差分投稿します。
    *   `json`: デバッグ用に、処理結果の内部構造をJSON形式で標準出力します。

### ビルドとセットアップ

```sh
$ go build .
```

このプログラムは、`mergeRss`, `diffRss` のように、実行されたファイル名（シンボリックリンク名）で振る舞いを変えることもできます。
最初にシンボリックリンクを作成するために `-s` オプションを実行してください。

```sh
$ ./feed2cli -s
```

### 使い方

#### 1. フィードのマージ

はてなブックマークの検索結果など、複数のフィードをマージします。

```sh
$ # ITカテゴリと総合カテゴリのホットエントリをマージする
$ echo "$(curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss')" \
       "$(curl -sL 'http://b.hatena.ne.jp/hotentry.rss')" \
  | ./mergeRss > merged_feed.rss
```

#### 2. はてなブックマーク情報のフィルタリングと出力

`-f hatena_bookmark` フィルタを使用すると、フィードの各エントリにはてなブックマークのブックマーク数とコメントが付与されます。

##### RSSとして出力

フィルタリング結果をRSSとして出力します。Descriptionに情報が追記されます。

```sh
$ # ITカテゴリのホットエントリを取得し、はてなブックマーク情報を付与したRSSを出力
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark > filtered_feed.rss
```

##### JSONとして出力（デバッグ用）

`-o json` を使うと、フィルタが生成した内部データ構造を直接確認できます。

```sh
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark -o json | jq .
```

#### 3. Slackへの通知

環境変数 `XOXB` (Slack Bot Token) と `SLACK_CHANNEL` (投稿先チャンネルID) の設定が必要です。

##### シンプルな通知 (`-o slack`)

フィードの内容を整形してSlackに投稿します。フィルタと組み合わせることで、はてなブックマークの件数などをメッセージに含めることもできます。

```sh
$ # フィルタなしで、フィードの内容をそのまま通知
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' | ./feed2cli -o slack

$ # フィルタを適用し、ブックマーク件数を含めて通知
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' | ./feed2cli -f hatena_bookmark -o slack
```

##### はてなコメントのスレッド通知 (`-o hatena`)

はてなブックマークのコメントを、Slackのスレッドに差分投稿するための専用機能です。**必ず `-f hatena_bookmark` と組み合わせて使用してください。**

```sh
$ # ITホットエントリにはてな情報を付与し、Slackのスレッドにコメントを投稿
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark -o hatena
```

このコマンドを実行すると、以下の処理が行われます。

1.  各エントリがSlackに投稿されます（初回のみ）。
2.  その投稿のスレッドに、はてなブックマークのコメントが投稿されます。
3.  再度同じコマンドを実行すると、新しく付いたコメント（差分）のみがスレッドに追加されます。