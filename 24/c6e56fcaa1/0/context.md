# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# SEO改善プラン — Meow Programming Language サイト

## Context

サイト `https://135yshr.github.io/meow/` が検索エンジンでトップに表示されない。調査の結果、以下の重大な問題が見つかった。

---

## 発見された問題

### 致命的 (Critical)

1. **`public/` が `localhost:1313` のURLで生成されている**
   - `sitemap.xml`、OGタグ、全てのURLが `http://localhost:1313/meow/` を指している
   - 原因: `public/` がローカル開発サーバーで生成され、そのままコミットされている
   - CI（`hugo.yml`）では `--baseURL` で正しいURLを使っているが、コミット済みの `public/` は開発用
   - **対策**: `public/` をgitignoreに追加（CIで毎回生成されるため不要）
...

### Prompt 2

create pr

