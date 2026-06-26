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

### CLI（非対話）

TUI を介さずにリポジトリ/Organization を登録・参照できます（スキルやスクリプトから利用）。

```sh
bako project list [--json]                 # プロジェクト一覧（slug / name）

# リポジトリ/Organization を登録（同じ owner/name は上書き更新）
bako repo add --project <slug|name> --owner <owner> --repo <name> [--url <url>] \
  [--overview <text> | --overview-file <path|-> ]
bako repo add --project <slug|name> --owner <owner> --org [--overview ...]

# 概要を標準入力から渡す例
echo "## 概要..." | bako repo add --project apex-rewrite --owner ystsbry --repo bako --overview-file -

bako repo list --project <slug|name> [--json]   # 登録済みリポジトリの確認
```

`--project` には slug でも表示名でも渡せます（bako 側で解決）。

### TUI の操作

- **プロジェクト一覧**: `↑/↓` 移動 · `enter` 開く · `n` 新規 · `e` 編集 · `d` 削除 · `q` 終了
- **プロジェクト詳細**: `tab` で PBI / Repo タブ切替 · `n` 新規 · `e` 編集 · `d` 削除 · `esc` 戻る
  - PBI タブでは `s` で状態を Todo → Progress → Done と切替
- **入力フォーム**: `tab` でフィールド移動 · `ctrl+s` 保存 · `esc` キャンセル
  - 状態 / 種別フィールドは `←/→` で切替
  - PBI 内容やリポジトリ概要はテキストエリアに貼り付け可能

## スキル / プラグイン（Claude Code・Codex）

`plugin/` 配下に、リポジトリ説明を bako に登録するためのスキルを同梱しています。

| スキル | 役割 |
| ------ | ---- |
| `register-repo` | 対象リポジトリを調査して日本語の概要を生成し、`bako repo add` で bako プロジェクトに登録する |

インストール:

```sh
make install            # bako CLI を PATH へ（スキルが内部で呼ぶ）
make install-plugin     # Claude Code: ~/.claude/skills/bako に登録（/bako:register-repo）
make install-codex-skills   # Codex CLI: ~/.agents/skills に登録（$register-repo）
```

使い方:

```
/bako:register-repo --project <slug|name> --repo <owner/name> [--path <REPO_DIR>] [--org]   # Claude Code
$register-repo --project <slug|name> --repo <owner/name> [--path <REPO_DIR>] [--org]         # Codex CLI
```

`--path` にローカルのリポジトリを渡すと、その内容（README・構成・依存）を調査して概要を作成します。
登録先プロジェクトは事前に TUI で作成しておく必要があります。

## 今後の予定

- CLI 経由での PBI ステータス取得・内容/タイトル出力
- PBI を読むためのスキル

## 開発

```sh
make test   # go test ./...
make vet
make fmt
```
