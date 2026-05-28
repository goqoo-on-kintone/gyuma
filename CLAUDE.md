# CLAUDE.md

このファイルはAI（Claude）がこのリポジトリを理解・操作する際のガイドです。

## プロジェクト概要

**Gyuma OAuth** は kintone 向けの OAuth 2.0 クライアントライブラリ。
API としての利用と、CLI ツールとしての利用の両方をサポートする。

### Go 版への移行

現在、Node.js/TypeScript 版から Go 版への全面書き直しを進めている。

- **設計書**: [docs/go-rewrite-design.md](docs/go-rewrite-design.md)
- **TODO**: [docs/TODO.md](docs/TODO.md)
- **Node.js 版アーカイブ**: `nodejs-archive` ブランチ

## 技術スタック（Go 版）

- **言語**: Go
- **主要パッケージ**:
  - `net/http` - HTTPS サーバー / OAuth コールバック
  - `crypto/x509` - 自己署名証明書生成
  - `gopkg.in/ini.v1` - クレデンシャルファイル（INI 形式）のパース

## ディレクトリ構成（Go 版）

```
gyuma/
├── cmd/
│   └── gyuma/
│       └── main.go           # CLI エントリポイント
├── internal/
│   ├── auth/
│   │   ├── server.go         # HTTPS サーバー / OAuth フロー
│   │   └── token.go          # トークンキャッシュ読み書き
│   ├── config/
│   │   ├── credentials.go    # クレデンシャル読み込み（優先順位制御）
│   │   └── paths.go          # 設定ディレクトリパス
│   ├── browser/
│   │   └── open.go           # ブラウザ起動（OS別）
│   └── cert/
│       ├── cert.go           # 自己署名証明書生成
│       └── mkcert.go         # mkcert 連携（検出・実行）
├── docs/
│   ├── go-rewrite-design.md  # 設計書
│   └── TODO.md               # 実装 TODO リスト
├── src/                      # Node.js 版（deprecated・後で削除予定）
├── go.mod
├── go.sum
├── Makefile
└── CLAUDE.md
```

## ビルド・実行

```bash
# ビルド
make build

# クロスコンパイル（全プラットフォーム）
make build-all

# ヘルプ表示
./bin/gyuma --help
```

## CLI インターフェース

```
gyuma [options]                    # OAuth トークン取得
gyuma setup-cert [options]         # mkcert 証明書のセットアップ

# 主要オプション
  -d, --domain           kintone ドメイン名（必須）
  -S, --scope            OAuth2 Scope（必須）
  -i, --client-id        OAuth2 Client ID
  -s, --client-secret    OAuth2 Client Secret
  --refresh-token        リフレッシュトークンを保存・利用
  --quiet                警告メッセージを抑制
```

## ファイル保存先（Go 版）

- 設定ルート: `~/.config/gyuma/`
- トークン: `~/.config/gyuma/tokens.json`（ドメインをキーとした JSON）
- クレデンシャル: `~/.config/gyuma/credentials`（INI 形式・プレーンテキスト）
- 証明書: `~/.config/gyuma/certs/`

## 開発ブランチ運用

- `main` - リリースブランチ
- `feature/go-rewrite` - Go 版実装ブランチ（現在の作業ブランチ）
- `nodejs-archive` - Node.js 版のコードを保存（参照用・変更しない）

## 注意事項

- Go 版の実装は `feature/go-rewrite` ブランチで行う
- SSL証明書（`ssl/` 配下）は `.gitignore` に含める
- `bin/` ディレクトリはビルド出力先で `.gitignore` に含める
