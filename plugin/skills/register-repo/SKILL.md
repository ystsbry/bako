---
name: register-repo
description: bako のプロジェクトに GitHub リポジトリ（または Organization）の説明（概要）を登録する。対象リポジトリを調査して日本語の概要を生成し、`bako repo add` で保存する。「リポジトリ説明を登録して」「bako にこのリポジトリを登録」「register-repo」などと言われたとき、またはリポジトリのパス/スラッグと登録先プロジェクトを渡されたときに使う。
---

# register-repo

あなたは、対象の GitHub リポジトリを調査し、その**概要（説明）**を簡潔な日本語のマークダウンにまとめて、bako のプロジェクトへ登録するアシスタントです。登録には bako の CLI を使います。

## 入力

```
/bako:register-repo --project <slug|name> --repo <owner/name> [--path <REPO_DIR>] [--org]   # Claude Code (plugin)
$register-repo --project <slug|name> --repo <owner/name> [--path <REPO_DIR>] [--org]         # OpenAI Codex CLI
```

- `--project <slug|name>` (必須): 登録先の bako プロジェクト。スラッグでも表示名でも可。
- `--repo <owner/name>` (必須): 対象リポジトリ。`owner/name` 形式（例: `ystsbry/bako`）。
  - `--org` を付けた場合は Organization の登録とみなし、`<owner>` のみ（`name` は不要）。
- `--path <REPO_DIR>` (任意・推奨): ローカルにクローン済みのリポジトリのパス。
  与えられた場合は**必ずこのディレクトリを調査**して概要を書く。リポジトリは**参照専用**で、ファイルを変更してはならない。
- 引数を伴わずチャットでリポジトリの情報・概要文を直接渡された場合は、その内容を対象として扱う。

`--path` が無く、対象がカレントリポジトリ（または会話の文脈で特定済み）の場合は、それを調査対象とみなしてよい。判断できない場合はユーザーに確認する。

## 前提

- `bako` CLI が PATH 上にあること（`bako version` で確認できる）。無ければビルド/インストールを促す（`make install` もしくは `go build -o bin/bako ./cmd/bako` の上で `./bin/bako` を使う）。
- 登録先プロジェクトは**既に存在**していること（プロジェクトの作成は bako の TUI で行う）。

## 手順

1. **CLI とプロジェクトの存在確認**
   - `bako version` で CLI を確認する。
   - `bako project list --json` を実行し、`--project` で指定された値が存在するか確認する（スラッグ・表示名のどちらでも `bako` 側が解決する）。
     見つからない場合は、一覧を提示して「TUI でプロジェクトを作成してください」と案内し、**登録は行わない**。

2. **リポジトリの調査**（概要を具体化するため）
   - `--path` がある場合は、そのディレクトリで以下を確認する:
     - `README*` / `docs/` … リポジトリの目的・使い方
     - 構成ファイル（`go.mod` / `package.json` / `pyproject.toml` / `Cargo.toml` など）… 言語・主要依存
     - トップレベルのディレクトリ構成 … 役割の概観（`cmd/` `internal/` `src/` など）
     - ビルド/実行方法（`Makefile` / `scripts/` / CI 設定）
   - `--path` が無い場合は、会話の文脈や既知の情報から書く。情報が不足し推測になる場合は、**断定を避け**「（要確認）」を添えるか、ユーザーに不足情報を尋ねる。

3. **概要（説明）の作成**
   - **日本語**のマークダウンで、簡潔に（おおむね 10〜40 行程度）。見出し・箇条書きを適切に使う。
   - 推奨する構成（過不足は対象に合わせて調整）:
     ```
     ## 概要
     - 一言で何のリポジトリか（目的・解決する課題）

     ## 技術スタック
     - 言語 / 主要フレームワーク・ライブラリ

     ## 構成
     - 主要ディレクトリ・モジュールとその役割

     ## 使い方 / 実行
     - ビルド・起動・主要コマンド（分かる範囲で）

     ## 備考
     - 関連リポジトリ、注意点、未確認事項（あれば）
     ```
   - 事実に基づいて書く。確認できない点を断定しない。

4. **登録（bako CLI 呼び出し）**
   - 生成した概要マークダウンを**標準入力経由**で渡し、`--overview-file -` で受け取らせる（引用やエスケープの事故を避けるため、シェル文字列に直接埋め込まない）。
   - リポジトリの場合:
     ```sh
     printf '%s' "$OVERVIEW_MARKDOWN" | bako repo add \
       --project "<slug|name>" --owner "<owner>" --repo "<name>" --overview-file -
     ```
   - Organization の場合（`--org`、`--repo` は不要）:
     ```sh
     printf '%s' "$OVERVIEW_MARKDOWN" | bako repo add \
       --project "<slug|name>" --owner "<owner>" --org --overview-file -
     ```
   - URL を明示したい場合のみ `--url <url>` を付ける（省略時は `github.com/owner[/name]` が自動補完される）。
   - 同じ owner/name を再登録すると**上書き更新**になる（重複は作られない）。

5. **確認**
   - `bako repo list --project "<slug|name>"` を実行して登録結果を確認し、登録した owner/name・URL と概要の要点をユーザーへ報告する。

## 出力

- 登録した内容（プロジェクト / 対象リポジトリ / URL）と、生成した概要マークダウンの要約を日本語で報告する。
- 登録できなかった場合（プロジェクト不在・CLI 不在・情報不足）は、理由と次のアクション（TUI でのプロジェクト作成、不足情報の確認など）を明示する。

## 注意

- 対象リポジトリのファイルは**読むだけ**。変更・コミットはしない。
- 概要の保存先は bako のローカルストレージ（既定 `~/.bako/<project>/repo/<owner>__<name>.md`、`BAKO_HOME` で変更可）。スキルから直接このファイルを書かず、必ず `bako repo add` を経由する（スラッグ生成・frontmatter 整形を CLI に委ねるため）。
