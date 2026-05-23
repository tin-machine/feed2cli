# feed2cli

### 概要

RSS/Atomフィードをマージ、差分、フィルタリングするためのCLIツールです。
Plaggerにインスパイアされており、Unixのパイプライン思想に沿って、手軽にフィードを加工することを目指しています。

例えば、はてなブックマークの複数キーワードの検索結果を追いかける際に、重複するエントリをマージしたり、各エントリにはてなブックマークのコメントを付与したり、その結果をSlackに通知したりできます。

### 主な機能

*   **merge**: 複数のフィードをマージし、URLを正規化したうえで重複エントリを排除します。
*   **diff**: 2つのフィードを比較し、URLを正規化したうえで新しいフィードにのみ存在するエントリを抽出します。
*   **filter**: `-f`フラグで指定されたフィルタを適用し、フィードの各エントリに追加情報を付与します。
    *   `hatena_bookmark`: はてなブックマークのブックマーク数とコメントを取得します。
*   **output**: `-o`フラグで指定された形式で結果を出力します。
    *   `(デフォルト)`: RSS形式で標準出力します。
    *   `atom`: Atom形式で標準出力します。
    *   `slack`: Slackに汎用的なメッセージを通知します。
    *   `hatena`: はてなブックマークのコメントをSlackのスレッドに差分投稿します。
    *   `digest`: 指定時間内のエントリをMarkdown digestとして標準出力します。
    *   `lint`: 入力されたRSS/Atomがparse可能か検証します。
    *   `json`: デバッグ用に、処理結果の内部構造をJSON形式で標準出力します。
    *   `jsonl`: `FeedItem`を1行1JSONで標準出力します。

重複排除のURL正規化では、`utm_*`, `fbclid`, `gclid` などのtracking query、fragment、host/schemeのcase、末尾slash、`m.` host、末尾 `/amp` を比較用keyから取り除きます。出力するURL自体は元のfeed itemのlinkを維持します。

### Unix pipeline 方針

feed2cli は、RSS/Atom を標準入力で受け取り、通常の加工結果を標準出力へ返す CLI として扱います。

```text
RSS/Atom stdin
  -> internal FeedDocument / FeedItem
  -> optional FeedItemStage chain
  -> RSS / Atom / Markdown / JSON / JSONL / lint report stdout
```

基本方針は次の通りです。

*   `merge`, `diff`, デフォルト出力は RSS を標準出力へ返します。
*   `atom` は Atom を標準出力へ返します。
*   `json` / `jsonl` は構造化データを標準出力へ返します。
*   `digest` は Markdown を標準出力へ返します。
*   `lint` は検証結果を標準出力へ返し、壊れた feed があれば終了コード1にします。
*   `slack` / `hatena` のような副作用を持つ出力は、明示指定された場合だけ実行します。

内部処理は `FeedItemStage` として段階的に分けられるようにしています。現時点では組み込み stage の境界だけを用意し、外部 plugin 化や設定ファイルからの組み立ては後続作業に回しています。

### JSONL intermediate

`-o jsonl` は、1行1エントリの安定した中間形式を出力します。各行には `schema_version: "feed2cli.feed_item.v1"` が入り、正規化前 URL は `url`、正規化後 URL は `normalized_url`、source は `source`、score や enrich 結果は `metadata` や専用 field に入ります。

```sh
$ ./feed2cli -url 'https://example.com/feed.xml' -o jsonl \
  | jq -c 'select(.normalized_url | contains("example.com"))' \
  | ./feed2cli -input jsonl -o atom
```

`-input jsonl` は `feed2cli.feed_item.v1` の行を読み込みます。移行用に、古い `FeedItem` 直出し JSONL も読み込めますが、新しく保存する archive や外部 plugin との接続には schema version 付きの形式を使います。

### Pipeline Config

`-config` で input、stage chain、output を JSON ファイルから指定できます。初期実装は JSON のみです。YAML 対応は、同じ構造体に parser を足す形で後続対応できます。

```json
{
  "input": {
    "format": "feed",
    "urls": ["https://example.com/feed.xml"]
  },
  "stages": [
    {"type": "normalize"},
    {"type": "merge"},
    {"type": "keyword_filter", "include": ["go"], "exclude": ["dog"]},
    {"type": "domain_filter", "include": ["example.com"]},
    {"type": "time_window", "since": "24h"},
    {"type": "hotness_score"},
    {"type": "rank", "by": "hotness"}
  ],
  "output": {
    "type": "jsonl"
  }
}
```

```sh
$ ./feed2cli -config pipeline.json
```

利用できる stage は `normalize`, `merge`, `hatena_bookmark`, `keyword_filter`, `domain_filter`, `time_window`, `hotness_score`, `min_hotness`, `fav_user`, `rank`, `source_label`, `tag`, `ogp`, `content`, `summary`, `plugin` です。既存 CLI flag は config の後段に追加処理として併用できます。

`plugin` stage は、外部 command と JSONL で接続します。feed2cli は `feed2cli.feed_item.v1` JSONL を stdin に渡し、plugin は同じ schema の JSONL を stdout に返します。stderr は診断用で、非ゼロ終了または timeout の場合に error message へ含めます。

```json
{
  "stages": [
    {
      "type": "plugin",
      "command": "/usr/local/bin/feed2cli-my-filter",
      "args": ["--keep", "go"],
      "timeout": "10s"
    }
  ],
  "output": {"type": "jsonl"}
}
```

plugin 実行では shell を挟まないため、pipe や redirect が必要な場合は小さい wrapper script を command として指定します。

### Explain

`-explain` は通常の output を実行せず、入力 item ごとに残す/落とす判定理由を JSONL で出します。schema は `feed2cli.explain.v1` です。

```sh
$ ./feed2cli -url 'https://example.com/feed.xml' \
    -include-keyword go \
    -exclude-keyword dog \
    -include-domain example.com \
    -o jsonl \
    -explain
```

`keyword_filter`, `domain_filter`, `time_window`, `hotness_score`, `min_hotness`, `fav_user`, `rank` の判定理由を出します。pipeline config の stage は順にシミュレーションし、各 record の `stages` に stage ごとの結果を入れます。

`source_label`, `tag`, `summary` は副作用なしでシミュレーションします。`ogp`, `content`, `hatena_bookmark`, `plugin` は外部 HTTP や外部 command を実行せず、dry-run の理由だけを出します。Slack への実投稿確認には `-slack-dry-run` を使います。

### ビルドとセットアップ

Go 1.26.3 以降を前提にしています。

```sh
$ go build .
```

### リリース

リリースバージョンは Git tag を正とします。`v0.0.17` のような tag を push すると GitHub Actions がリリース用バイナリをビルドし、GitHub Release を作成します。

```sh
$ git checkout main
$ git pull
$ git tag -a v0.0.18 -m "v0.0.18"
$ git push origin v0.0.18
```

このプログラムは、`mergeRss`, `diffRss` のように、実行されたファイル名（シンボリックリンク名）で振る舞いを変えることもできます。
最初にシンボリックリンクを作成するために `-s` オプションを実行してください。

```sh
$ ./feed2cli -s
```

フィード URL は `-url` で直接指定できます。複数回指定した場合は、それぞれ取得して同じ入力として扱います。
標準入力と `-url` は併用でき、その場合は標準入力で渡した feed と URL 取得した feed を合わせて処理します。

```sh
$ ./feed2cli -url 'https://example.com/feed.xml' -o atom

$ ./feed2cli \
    -url 'https://example.com/a.xml' \
    -url 'https://example.com/b.xml' \
    -o merge > merged.rss

$ cat local.rss | ./feed2cli -url 'https://example.com/feed.xml' -o jsonl

$ ./feed2cli -url 'https://example.com/feed.xml' -o jsonl \
  | ./feed2cli -input jsonl -include-keyword go -o digest
```

`-enrich` は複数回指定でき、出力前に `FeedItem` へ情報を付与します。

```sh
$ ./feed2cli -url 'https://example.com/feed.xml' \
    -enrich source_label \
    -enrich ogp \
    -enrich content \
    -enrich summary \
    -o jsonl
```

利用できる組み込み enrich は次の通りです。

*   `source_label`: feed source 名を `source` / `metadata.source_label` に入れます。
*   `tag`: はてなブックマークコメントの tag を category に統合します。
*   `ogp`: 記事 URL から `og:title`, `og:description`, `og:image`, `og:site_name` を取得します。
*   `content`: 記事 URL から本文候補を取得し、`content` / `metadata.content_text` に入れます。
*   `summary`: 外部APIを呼ばない抽出型 summary を作ります。将来 local LLM に差し替えられる `Summarizer` 境界を使っています。

フィルタとランキングも `FeedItemStage` として適用できます。

```sh
$ ./feed2cli -url 'https://example.com/feed.xml' \
    -include-keyword go \
    -exclude-keyword dog \
    -include-domain example.com \
    -since 24h \
    -rank hotness \
    -o jsonl
```

主な指定は次の通りです。

*   `-include-keyword`: title / description / content / category に語を含む item を残します。
*   `-exclude-keyword`: 語を含む item を落とします。
*   `-min-keyword-score`: `-include-keyword` の一致数しきい値です。
*   `-include-domain` / `-exclude-domain`: URL domain で残す/落とす条件です。
*   `-since`: `24h` などの時間窓で item を絞ります。
*   `-fav-user`: 指定したはてなブックマーク user のコメントがある item を残します。
*   `-rank hotness`: はてブ数、コメント数、鮮度から計算した `metadata.hotness_score` で並べます。
*   `-min-hotness`: hotness score のしきい値です。

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

環境変数 `FEED2CLI_SLACK_TOKEN` (Slack Bot Token) と `FEED2CLI_SLACK_CHANNEL` (投稿先チャンネルID) の設定が必要です。
互換性のため `XOXB` / `SLACK_CHANNEL` も読みます。CLI から `-slack-channel` で投稿先を上書きできます。

実投稿では事前に `auth.test` と channel validation を行います。`#channel-name` も指定できますが、Slack API scope によっては name 解決できないため、運用では channel ID (`C...` / `G...`) を推奨します。検証だけ行いたい場合は `-slack-dry-run` を使います。

##### シンプルな通知 (`-o slack`)

フィードの内容を整形してSlackに投稿します。フィルタと組み合わせることで、はてなブックマークの件数などをメッセージに含めることもできます。

```sh
$ # フィルタなしで、フィードの内容をそのまま通知
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' | ./feed2cli -o slack

$ # フィルタを適用し、ブックマーク件数を含めて通知
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' | ./feed2cli -f hatena_bookmark -o slack

$ # 投稿せず、Slackに送る予定の内容だけ確認
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -o slack -slack-dry-run -slack-channel C0123456789
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

状態管理はデフォルトで `hatena_state.json` を使います。SQLiteを使う場合は `-state-backend sqlite` と `-state-path` を指定します。
state には記事URLやSlack thread timestampが入るため、公開repositoryには含めないでください。

```sh
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark -o hatena \
      -state-backend sqlite \
      -state-path ./hatena_state.sqlite
```

同じ設定は `FEED2CLI_STATE_BACKEND=sqlite` と `FEED2CLI_STATE_PATH=./hatena_state.sqlite` でも指定できます。

`-o hatena` でも `-slack-dry-run` を指定できます。この場合は state を読み、親投稿とコメント投稿の予定件数を表示しますが、Slack投稿も state 保存も行いません。

```sh
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -f hatena_bookmark -o hatena \
      -slack-dry-run \
      -state-backend sqlite \
      -state-path ./hatena_state.sqlite
```

#### 4. Digest出力

`-o digest` は、指定時間内のフィードエントリをMarkdownでまとめます。

```sh
$ curl -sL 'http://b.hatena.ne.jp/hotentry/it.rss' \
  | ./feed2cli -o digest -digest-window 24h -digest-title "IT hotentry"
```

`-digest-window 0` を指定すると、入力された全エントリを対象にします。

#### 5. Feed lint

`-o lint` は、入力されたRSS/Atomをparseして件数とエラーを表示します。壊れたfeedが含まれる場合は終了コード1になります。

```sh
$ cat feeds.xml | ./feed2cli -o lint
feeds: total=2 valid=2 invalid=0
```

#### 6. Slack integration test

通常のテストはfake posterでSlack APIへ通信しません。実API疎通を確認する場合だけ、明示的にopt-inします。

Codexが動作確認でSlackへ通信する場合は、`CODEX_SLACK_` prefixの環境変数だけを使います。

```sh
$ FEED2CLI_SLACK_INTEGRATION=1 \
  CODEX_SLACK_BOT_TOKEN=xoxb-... \
  CODEX_SLACK_CHANNEL=C0123456789 \
  go test ./... -run TestSlackIntegrationOptIn
```

`CODEX_SLACK_CHANNEL` が未設定の場合は、`CODEX_SLACK_TEST_CHANNEL_NAME` を使って既存 channel を探し、見つからなければテスト用 channel を作成して投稿します。
