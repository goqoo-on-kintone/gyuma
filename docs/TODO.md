# Go 書き直し TODO

> 実装は `feature/go-rewrite` ブランチで行う

## Phase 1: Go バイナリの実装

### 1.1 プロジェクト初期化

- [x] Go モジュール初期化（`go mod init`）
- [x] ディレクトリ構成の作成（`cmd/`, `internal/`）
- [x] Makefile 作成（ビルド・クロスコンパイル用）
- [x] Node.js 版のファイルを削除（`nodejs-archive` ブランチに保存済み）

### 1.2 設定・パス管理（internal/config）

- [x] `paths.go` - 設定ディレクトリパス（`~/.config/gyuma/`）
- [x] `credentials.go` - クレデンシャル読み込み（優先順位制御）
  - CLI オプション → 環境変数 → ファイル → プロンプト

### 1.3 証明書管理（internal/cert）

- [x] `cert.go` - 自己署名証明書の生成
- [x] `mkcert.go` - mkcert 連携（検出・実行）

### 1.4 認証フロー（internal/auth）

- [x] `token.go` - トークンキャッシュの読み書き（`tokens.json`）
- [x] `server.go` - HTTPS サーバー / OAuth コールバック処理

### 1.5 ブラウザ起動（internal/browser）

- [x] `open.go` - OS 別ブラウザ起動

### 1.6 CLI エントリポイント（cmd/gyuma）

- [x] `main.go` - CLI 引数パース、メインフロー
- [x] サブコマンド `setup-cert` の実装

### 1.7 テスト・動作確認

- [x] 単体テスト（各パッケージ）
- [x] 結合テスト（実際の kintone 環境で OAuth フロー確認）

---

## Phase 2: npm パッケージの整備

### 2.1 JS ラッパー（npm/gyuma）

- [ ] `package.json` 作成
- [ ] `lib/index.ts` - Go バイナリを呼び出すラッパー
- [ ] `lib/resolve-binary.ts` - プラットフォーム別バイナリパス解決
- [ ] `bin/gyuma.js` - CLI シム

### 2.2 プラットフォーム別バイナリパッケージ

- [ ] `gyuma-darwin-arm64` パッケージ
- [ ] `gyuma-darwin-x64` パッケージ
- [ ] `gyuma-linux-x64` パッケージ
- [ ] `gyuma-linux-arm64` パッケージ
- [ ] `gyuma-win32-x64` パッケージ

---

## Phase 3: リリース準備

### 3.1 CI/CD

- [ ] GitHub Actions ワークフロー（クロスコンパイル）
- [ ] npm publish 自動化

### 3.2 ドキュメント

- [ ] README.md 更新（Go 版の使い方）
- [ ] CLAUDE.md 更新（Go 版の構成を反映）

### 3.3 Node.js 版の deprecated 化

- [ ] package.json に deprecated 明記
- [ ] README に移行案内を記載

---

## 補足

- **nodejs-archive ブランチ**: Node.js 版のコードを保存
- **設計書**: [docs/go-rewrite-design.md](./go-rewrite-design.md)
