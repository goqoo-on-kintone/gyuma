# CLAUDE.md

このファイルはAI（Claude）がこのリポジトリを理解・操作する際のガイドです。

## プロジェクト概要

**Gyuma OAuth** は kintone 向けの OAuth 2.0 クライアントライブラリ（Node.js / TypeScript）。  
API としての利用と、CLI ツールとしての利用の両方をサポートする。

## 技術スタック

- **言語**: TypeScript
- **ランタイム**: Node.js
- **主要依存ライブラリ**:
  - `express` - OAuth コールバック受信用 HTTPS サーバー
  - `node-fetch` - kintone トークンエンドポイントへの HTTP リクエスト
  - `proxy-agent` - プロキシ対応
  - `minimist` - CLI 引数パース
  - `date-fns` / `moment` - トークン有効期限の計算
  - `opener` - ブラウザ自動起動
  - `qs` - クエリストリングのシリアライズ

## ディレクトリ構成

```
gyuma/
├── src/
│   ├── index.ts          # エントリポイント（named export）
│   ├── main.ts           # gyuma() メイン関数 / トークンキャッシュロジック
│   ├── server.ts         # HTTPS サーバー / OAuth 認可フロー処理
│   ├── agent.ts          # HTTP エージェント生成（プロキシ / クライアント証明書）
│   ├── cli.ts            # CLI エントリポイント
│   ├── config-file-io.ts # トークン・クレデンシャルのファイル I/O
│   ├── encrypt.ts        # クレデンシャルの暗号化・復号
│   ├── input-password.ts # インタラクティブなパスワード入力
│   ├── createCertificate.ts # 自己署名証明書の生成
│   ├── constants.ts      # 設定ディレクトリパス等の定数
│   └── types.ts          # TypeScript 型定義
├── ssl/                  # 自動生成される自己署名証明書の格納先
├── test/                 # テストコード
├── package.json
└── tsconfig.json
```

## アーキテクチャ概要

### OAuth フロー

1. `gyuma(argv)` が呼ばれる
2. キャッシュ済みトークンの有無・scope 一致・有効期限をチェック
3. 有効なトークンがあればそのまま返す（ブラウザ起動なし）
4. なければ `server()` を起動：
   - ローカルに HTTPS サーバー（デフォルト `https://localhost:3000`）を立ち上げる
   - ブラウザを自動起動して kintone の OAuth 認可ページへリダイレクト
   - コールバック（`/oauth2callback`）でアクセストークンを取得
   - CSRF 対策として `state` パラメータを検証
5. 取得したトークンを `~/.config/gyuma/<domain>/token.json` にキャッシュ
6. クレデンシャル（client_id / client_secret）は暗号化して `~/.config/gyuma/<domain>/credentials` に保存

### ファイル保存先

- 設定ルート: `~/.config/gyuma/`
- トークン: `~/.config/gyuma/<domain>/token.json`
- クレデンシャル: `~/.config/gyuma/<domain>/credentials`（暗号化済み）

## 型定義（types.ts）

```typescript
type Argv = {
  domain: string
  scope: string
  password?: string
  client_id?: string
  client_secret?: string
  port?: number
  proxy?: ProxyOption
  pfx?: PfxOption
  noprompt?: boolean
}

type Token = {
  expiry: string
  refresh_token?: string  // 現在は保存しない（writeToken で削除）
  access_token: string
  scope: string
}
```

## 既知の TODO / 技術的負債

- `config-file-io.ts` にコメントあり：クレデンシャルファイルを1ファイルに統合したい（現状はドメインごとにパスワードが必要）
- `server.ts` にコメントあり：ブラウザ自動起動しないマニュアルモードの追加
- `moment` と `date-fns` が混在している（要統一）
- SSL証明書は30日ごとに再生成される仕組みだが、`@ts-expect-error` を使っている箇所あり

## 開発ブランチ運用

- `main` - リリースブランチ
- `feature/design` - 設計ドキュメント作業ブランチ（現行の作業ブランチ）

## 注意事項

- **コードの変更はこのブランチでは行わない**。設計書の作成・更新のみを行う。
- SSL証明書（`ssl/` 配下）は `.gitignore` に含めること（自動生成されるため）
- PAT はチャット完了後に必ず Revoke すること
