# Gyuma

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](/README.md) | 日本語

kintone の OAuth 2.0 認証を行う CLI ツール。ローカル HTTPS サーバーを起動して OAuth コールバックを処理し、アクセストークンを出力します。

## インストール

### npm

```bash
npm install -g gyuma
```

グローバルインストールなしで使う場合：

```bash
npx gyuma -d example.cybozu.com -S "k:app_settings:read"
```

### Homebrew（macOS / Linux）

```bash
brew install goqoo-on-kintone/tap/gyuma
```

### go install

```bash
go install github.com/goqoo-on-kintone/gyuma/cmd/gyuma@latest
```

### バイナリダウンロード

[GitHub Releases](https://github.com/goqoo-on-kintone/gyuma/releases) からプラットフォーム別のバイナリをダウンロード：

- `gyuma_X.X.X_darwin_amd64.tar.gz`（macOS Intel）
- `gyuma_X.X.X_darwin_arm64.tar.gz`（macOS Apple Silicon）
- `gyuma_X.X.X_linux_amd64.tar.gz`（Linux x64）
- `gyuma_X.X.X_linux_arm64.tar.gz`（Linux ARM64）
- `gyuma_X.X.X_windows_amd64.zip`（Windows x64）
- `gyuma_X.X.X_windows_arm64.zip`（Windows ARM64）

### ソースからビルド

```bash
git clone https://github.com/goqoo-on-kintone/gyuma.git
cd gyuma
make build
# bin/gyuma が生成される
```

## 使い方

### アクセストークンの取得

```bash
# 基本的な使い方（対話モード）
gyuma -d example.cybozu.com -S "k:app_settings:read"

# クレデンシャルを指定
gyuma -d example.cybozu.com \
  -i YOUR_CLIENT_ID \
  -s YOUR_CLIENT_SECRET \
  -S "k:app_settings:read k:app_settings:write"

# リフレッシュトークンを使用
gyuma -d example.cybozu.com -S "k:app_settings:read" --refresh-token

# クレデンシャルを保存して次回以降に利用
gyuma -d example.cybozu.com -i YOUR_CLIENT_ID -s YOUR_CLIENT_SECRET \
  -S "k:app_settings:read" --save-credentials
```

### 証明書のセットアップ

ブラウザの SSL 警告なしでスムーズに OAuth フローを実行するには、mkcert 証明書を使用します：

```bash
# まず mkcert をインストール（https://github.com/FiloSottile/mkcert）
brew install mkcert
mkcert -install

# gyuma 用の証明書をセットアップ
gyuma setup-cert
```

## オプション

### メインコマンド

| オプション | 短縮形 | デフォルト | 説明 |
|-----------|-------|-----------|------|
| `--domain` | `-d` | | kintone ドメイン名（必須） |
| `--client-id` | `-i` | | OAuth2 クライアント ID |
| `--client-secret` | `-s` | | OAuth2 クライアントシークレット |
| `--scope` | `-S` | | OAuth2 スコープ（必須） |
| `--port` | `-P` | `3000` | ローカルサーバーのポート |
| `--refresh-token` | | `false` | リフレッシュトークンを保存・使用 |
| `--save-credentials` | | `false` | クレデンシャルをファイルに保存 |
| `--noprompt` | | `false` | 対話入力を無効化 |
| `--quiet` | | `false` | 警告メッセージを抑制 |
| `--version` | `-v` | | バージョンを表示 |
| `--help` | `-h` | | ヘルプを表示 |

### setup-cert サブコマンド

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--host` | `localhost` | 証明書のホスト名 |
| `--port` | `3000` | HTTPS サーバーのポート |
| `--force` | `false` | 既存の証明書を上書き |

## 設定

### クレデンシャルの優先順位

クレデンシャルは以下の順序で解決されます：

1. CLI オプション（`-i`、`-s`）
2. 環境変数（`GYUMA_CLIENT_ID`、`GYUMA_CLIENT_SECRET`）
3. クレデンシャルファイル（`~/.config/gyuma/credentials`）
4. 対話プロンプト（`--noprompt` でない場合）

### ファイルの保存先

| ファイル | パス | 説明 |
|---------|------|------|
| トークン | `~/.config/gyuma/tokens.json` | キャッシュされたアクセス/リフレッシュトークン |
| クレデンシャル | `~/.config/gyuma/credentials` | 保存されたクライアント認証情報（INI 形式） |
| 証明書 | `~/.config/gyuma/certs/` | ローカルサーバー用の SSL 証明書 |

### 環境変数

```bash
export GYUMA_CLIENT_ID="your_client_id"
export GYUMA_CLIENT_SECRET="your_client_secret"
```

## 他のツールとの連携

### シェルスクリプト

```bash
TOKEN=$(gyuma -d example.cybozu.com -S "k:app_settings:read")

curl -H "Authorization: Bearer $TOKEN" \
  "https://example.cybozu.com/k/v1/app/form/fields.json?app=1"
```

### ginue

[ginue](https://github.com/goqoo-on-kintone/ginue) で OAuth 認証を使う場合：

```bash
ginue pull --oauth
```

## kintone OAuth ドキュメント

- [How to add OAuth clients - English](https://kintone.dev/en/docs/common/authentication/how-to-add-oauth-clients/)
- [OAuthクライアントの使用 - 日本語](https://cybozu.dev/ja/kintone/docs/common/authentication/how-to-add-oauth-clients/)

## 開発

```bash
# ビルド
make build

# テスト
make test

# クロスコンパイル（全プラットフォーム）
make build-all
```

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) を参照してください。
