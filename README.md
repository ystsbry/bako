# bako

ローカルファースト（ローカルのマークダウンファイル）で、プロジェクト単位に
**PBI（Product Backlog Item）** と **GitHub リポジトリ / Organization** の情報を
まとめて管理するための TUI ツールです。

## できること（現状）

- プロジェクトごとに PBI を登録できる
- プロジェクトごとに GitHub のリポジトリ、もしくは Organization を登録できる
- リポジトリの概要（overview）を置いておける
- PBI ごとに **Todo / Progress / Done** の 3 状態を持てる
- PBI の内容は TUI 上のテキストエリアに貼り付け（コピペ）で入力できる

すべてのデータはローカルのマークダウンファイルとして保存されます。

## 保存形式

デフォルトでは `~/.bako/` 配下に保存します（`BAKO_HOME` 環境変数で変更可能）。

```
~/.bako/
  {project-slug}/
    project.md             # frontmatter に name、本文に説明
    pbi/
      {id}-{slug}.md        # frontmatter に title / status / created、本文に PBI 内容
    repo/
      {owner}__{name}.md    # frontmatter に kind / owner / name / url、本文に概要
```

PBI ファイルの例:

```markdown
---
title: ログイン機能の実装
status: progress
created: 2026-06-26
---

（ここに PBI の内容をコピペ）
```

リポジトリファイルの例（Organization の場合は `kind: org`、`name` は空）:

```markdown
---
kind: repo
owner: ystsbry
name: bako
url: https://github.com/ystsbry/bako
---

リポジトリの概要メモ。
```

## 使い方

```sh
make build
./bin/bako            # 引数なしで TUI を起動
./bin/bako home       # 保存ディレクトリを表示
./bin/bako version
```

### TUI の操作

- **プロジェクト一覧**: `↑/↓` 移動 · `enter` 開く · `n` 新規 · `e` 編集 · `d` 削除 · `q` 終了
- **プロジェクト詳細**: `tab` で PBI / Repo タブ切替 · `n` 新規 · `e` 編集 · `d` 削除 · `esc` 戻る
  - PBI タブでは `s` で状態を Todo → Progress → Done と切替
- **入力フォーム**: `tab` でフィールド移動 · `ctrl+s` 保存 · `esc` キャンセル
  - 状態 / 種別フィールドは `←/→` で切替
  - PBI 内容やリポジトリ概要はテキストエリアに貼り付け可能

## 今後の予定

- CLI 経由での PBI ステータス取得・内容/タイトル出力
- それらを読むためのスキル

## 開発

```sh
make test   # go test ./...
make vet
make fmt
```
