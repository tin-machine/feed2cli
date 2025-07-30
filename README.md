# feed2cli

### 概要

RSSフィードをマージ、差分、フィルタリングするためのCLIツールです。
Plaggerにインスパイアされており、パイプライン処理で手軽にフィードを加工することを目指しています。

例えば、はてなブックマークの複数キーワードの検索結果を追いかける際に、重複するエントリをマージしたり、各エントリにはてなブックマークのコメントを付与したりできます。

### 主な機能

*   **merge**: 複数のRSS/Atomフィードをマージし、重複エントリを排除します。
*   **diff**: 2つのフィードを比較し、新しいフィードにのみ存在するエントリを抽出します。
*   **filter**: フィードの各エントリに対して、追加情報を付与するフィルタを適用します。
*   **slack/hatena**: フィルタリングした結果をSlackに通知します。

### ビルドとセットアップ

```sh
$ go build .
```

このプログラムは、`mergeRss`, `diffRss` のように、実行されたファイル名（シンボリックリンク名）で振る舞いを変えます。
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

#### 2. はてなブックマーク情報のフィルタリング

`-f hatena_bookmark` フィルタを使用すると、フィードの各エントリにはてなブックマークのブックマーク数とコメントが付与されます。

```sh
$ # ITカテゴリのホットエントリを取得し、はてなブックマーク情報を付与する
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark > filtered_feed.rss
```

出力される `filtered_feed.rss` の各エントリのDescriptionには、以下のようにブックマーク数とコメントがHTML形式で追記されます。

```html
...元のDescription...
<p>Hatena Bookmark: <b>123</b></p>
<p><b>Comments:</b></p>
<ul>
  <li>user1 (2023/10/27 15:04:05): 面白い！</li>
  <li>user2 (2023/10/27 16:10:00): これはすごい。</li>
</ul>
```

#### 3. Slackへの通知

`-o hatena` オプションを使用すると、フィルタリングされた情報を元にSlackへ通知を送信できます。
この機能を利用するには、環境変数 `XOXB` (Slack Bot Token) と `SLACK_CHANNEL` (投稿先チャンネルID) の設定が必要です。

```sh
$ export XOXB="xoxb-your-slack-bot-token"
$ export SLACK_CHANNEL="C0123456789"

$ # ITホットエントリにはてな情報を付与し、Slackに通知する
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark -o hatena
```

このコマンドを実行すると、以下の処理が行われます。

1.  各エントリがSlackに投稿されます（初回のみ）。
2.  その投稿のスレッドに、はてなブックマークのコメントが投稿されます。
3.  再度同じコマンドを実行すると、新しく付いたコメント（差分）のみがスレッドに追加されます。

```
