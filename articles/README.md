# articles/

Zenn 形式の記事ドラフトを置くディレクトリです。

Zenn の GitHub 連携は `articles/<slug>.md` というレイアウトを前提にしているため、その慣習に合わせています。各ファイルの先頭には Zenn の front matter を書きます。

```yaml
---
title: "記事タイトル"
emoji: "🐱"
type: "tech" # tech（技術記事） または idea（アイデア記事）
topics: ["go", "compiler"] # 最大 5 つ
published: false
---
```

## `published: false` について

`published: false` は **未公開（下書き）** を意味します。Zenn の GitHub 連携を有効にしているリポジトリでも、この値が `false` である限り記事は公開されません。公開する準備ができたら `true` に変えてコミットします。

このディレクトリのドラフトはすべて `published: false` で追加しています。公開のタイミングは書き手が判断してください。

## Zenn の GitHub 連携について

このリポジトリで Zenn の GitHub 連携を有効にするかどうかは、リポジトリオーナーの判断に委ねられています。連携は設定されていない前提で、ここにあるのはあくまで Markdown のドラフトです。

連携せずに使う場合は、記事本文を Zenn のエディタや Qiita へコピーして投稿しても問題ありません。その場合 front matter は貼り付け先に合わせて調整してください。

## 記事を書くときの決まりごと

- 技術的な記述は、このリポジトリの実際のコードに基づいて書くこと。
- 記事に載せる `.nyan` のサンプルは、必ず `meow run` で実行して出力を確認すること。生成される Go コードを載せる場合は `meow transpile` の実出力をそのまま貼ること。
- 公式サイト（<https://meow.oreha.dev/>）、Playground（<https://meow.oreha.dev/playground/>）、GitHub リポジトリ（<https://github.com/135yshr/meow>）へのリンクを自然な位置に含めること。

## 記事一覧

| ファイル | タイトル | 状態 |
| --- | --- | --- |
| [`selfmade-lang-transpile-to-go.md`](./selfmade-lang-transpile-to-go.md) | Go にトランスパイルする自作言語の作り方 — lexer から native binary まで | 未公開 |
