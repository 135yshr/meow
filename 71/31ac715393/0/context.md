# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# 自動バージョン管理・リリースの導入

## Context

現在のmeowプロジェクトでは、マージ済みPRに `/release` コメントを手動で書くことでリリースをトリガーしている。これを **mainブランチへのマージ時に自動実行** する方式に変更し、バージョンタグ・CHANGELOG・GitHub Release（リリースノート付き）を自動生成する。

## 変更一覧

| ファイル | 操作 |
|---------|------|
| `.github/workflows/auto-release.yml` | **新規作成** — mainへのpush時にsemantic-release + GoReleaserを実行 |
| `.releaserc.json` | **更新** — `@semantic-release/changelog` と `@semantic-release/git` プラグインを追加 |
| `.github/workflows/goreleaser.yaml` | *...

### Prompt 2

# Smart Commit with Gitmoji

Execute the following steps non-interactively:

## Branch Management

- If currently on `main` or `master` branch, create and checkout a new feature branch with a descriptive name
- If on any other branch, proceed with commit on current branch (no branch creation)
- If no changes are detected, exit without doing anything

## Change Analysis & Commit

1. **Review all changes** using `git status` and `git diff --staged` (or `git diff` if nothing staged)
2. **Stage c...

### Prompt 3

create pr

### Prompt 4

Test Planに書かれた情報を確認して完了したら、チェックをつけてください

### Prompt 5

コードレビューが完了し、指摘がありました。
内容を確認して修正してください

### Prompt 6

# Smart Commit with Gitmoji

Execute the following steps non-interactively:

## Branch Management

- If currently on `main` or `master` branch, create and checkout a new feature branch with a descriptive name
- If on any other branch, proceed with commit on current branch (no branch creation)
- If no changes are detected, exit without doing anything

## Change Analysis & Commit

1. **Review all changes** using `git status` and `git diff --staged` (or `git diff` if nothing staged)
2. **Stage c...

