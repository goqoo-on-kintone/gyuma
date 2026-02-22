# Go 全面書き直し 設計書

## 背景・目的

現在の gyuma は Node.js / TypeScript 製のライブラリ・CLI ツールであり、利用するには Node.js ランタイムが必要。  
Go で書き直すことで以下を実現する：

- **単一バイナリ配布**: Node.js 不要。どの言語・環境からでも使いやすい
- **クロスコンパイル**: `GOOS` / `GOARCH` の指定だけで全プラットフォーム向けバイナリを生成
- **Node.js ライブラリとしての後方互換維持**: 既存ユーザーへの影響をゼロにする
- **シンプル化**: 暗号化機能を廃止し、クレデンシャル管理の責務を外部ツールに委譲

---

## Gyuma ファミリーのパッケージ構成

Gyuma を単一のツールとして閉じず、用途・利用者ごとに責務を分割したパッケージファミリーとして設計する。

### パッケージ一覧

| パッケージ | 言語 | 位置づけ | 主な利用者 |
|---|---|---|---|
| `@goqoo/gyuma-cli` | Go | CLIツール本体（Goバイナリ） | エンドユーザー（CLI） |
| `@goqoo/gyuma` | TypeScript | JSラッパー | Node.js CLIツール（Ginueなど）、Node.jsから使いたい人 |

### 依存関係

```
@goqoo/gyuma-cli（Goバイナリ）
CLIツール本体
・kintone OAuth フロー（golang.org/x/oauth2 を利用）
・クレデンシャル取得（優先順位に従う）
・HTTPS サーバー起動
・ブラウザ自動起動
・トークンキャッシュ
・access_token を stdout に出力して終了
        ↑
@goqoo/gyuma（JSラッパー）
・child_process で Goバイナリを呼ぶ
・access_token を stdout からキャプチャ
・Node.js の require()/import に対応
```

### 利用者ごとの使い方

**CLIエンドユーザー**  
`@goqoo/gyuma-cli` のGoバイナリを直接使う。

**Node.jsのCLIツール開発者（Ginueなど）**  
`@goqoo/gyuma` を `require()` する。内部でGoバイナリを呼ぶことを意識する必要はなく、`access_token` が返ってくればよい。

### リポジトリ構成

| リポジトリ | 内容 |
|---|---|
| `goqoo-on-kintone/gyuma` | `@goqoo/gyuma-cli`（Goバイナリ）+ `@goqoo/gyuma`（JSラッパー） |

---

## クレデンシャル管理の設計方針

### 基本思想

暗号化機能は **Gyuma から完全に廃止する**。

クレデンシャル（client_id / client_secret）の保護は、ユーザーが適切なツールで行う責務とする。  
これは `~/.aws/credentials` をプレーンテキストで保存する AWS SDK と同じ思想。  
1Password CLI や macOS Keychain などを使いたいユーザーは、Gyuma を介さず環境変数で渡せばよい。

### クレデンシャル取得の優先順位

以下の順で上から探し、最初に見つかったものを使う：

```
1. CLIオプション（--client-id, --client-secret）
2. 環境変数（GYUMA_CLIENT_ID, GYUMA_CLIENT_SECRET）
3. ~/.config/gyuma/credentials の該当ドメインセクション
4. インタラクティブな入力プロンプト（--noprompt なしの場合）
```

### 環境変数について

環境変数名はドメインによらず共通（`GYUMA_CLIENT_ID` / `GYUMA_CLIENT_SECRET`）。  
複数ドメインを使い分ける場合は、呼び出しごとに環境変数を切り替えて渡す：

```bash
# ドメインAにアクセス
GYUMA_CLIENT_ID=xxx GYUMA_CLIENT_SECRET=yyy gyuma --domain a.cybozu.com --scope "..."

# ドメインBにアクセス
GYUMA_CLIENT_ID=xxx GYUMA_CLIENT_SECRET=yyy gyuma --domain b.cybozu.com --scope "..."
```

### クレデンシャルファイル（プレーンテキスト）

AWS SDK の `~/.aws/credentials` と同様の INI 形式を採用する。ドメインをセクションキーとする。

**保存先**: `~/.config/gyuma/credentials`

```ini
[example.cybozu.com]
client_id     = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
client_secret = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

[another.cybozu.com]
client_id     = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
client_secret = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

ファイルのパーミッションは作成時に自動で `600` に設定する。

---

## ファイル構成

```
~/.config/gyuma/
├── credentials       # クレデンシャル（プレーンテキスト・任意）
└── tokens.json       # トークンキャッシュ（ドメインをキーとした JSON）
```

### tokens.json のフォーマット

```json
{
  "example.cybozu.com": {
    "access_token": "xxxxxxxx",
    "refresh_token": "xxxxxxxx",
    "scope": "k:app_settings:read k:app_settings:write",
    "expiry": "2024-01-01T00:00:00+09:00"
  },
  "another.cybozu.com": {
    "access_token": "xxxxxxxx",
    "scope": "k:app_settings:read",
    "expiry": "2024-01-01T00:00:00+09:00"
  }
}
```

> 現行は `~/.config/gyuma/<domain>/token.json` とドメインごとにディレクトリが分かれているが、  
> Go 版では `tokens.json` に統合する。Node.js 版からの移行時はトークンキャッシュの引き継ぎはできないため、初回のみ再認証が必要。

---

## リフレッシュトークンの扱い

現行はリフレッシュトークンを保存しない設計。  
Go 版では機能として実装するが、**デフォルトは無効・オプトインで有効化**する。

```
--refresh-token   リフレッシュトークンを保存し、アクセストークン期限切れ時に自動更新する
```

有効時のフロー：

```
1. アクセストークンが期限切れ
2. リフレッシュトークンが tokens.json に保存されていれば
   → トークンエンドポイントに refresh_token グラントで POST
   → 新しいアクセストークンを取得（ブラウザ起動なし）
3. リフレッシュトークンも期限切れ or 存在しない
   → 通常の認可フロー（ブラウザ起動）
```

---

## @goqoo/gyuma-cli（Goバイナリ）の設計

### ディレクトリ構成

```
gyuma/                              ← リポジトリルート（goqoo-on-kintone/gyuma）
├── cmd/
│   └── gyuma/
│       └── main.go                 # エントリポイント
├── internal/
│   ├── auth/
│   │   ├── server.go               # HTTPS サーバー / OAuth フロー
│   │   └── token.go                # トークンキャッシュ読み書き
│   ├── config/
│   │   ├── credentials.go          # クレデンシャル読み込み（優先順位制御）
│   │   └── paths.go                # 設定ディレクトリパス
│   ├── browser/
│   │   └── open.go                 # ブラウザ起動（OS別）
│   └── cert/
│       └── cert.go                 # 自己署名証明書生成
├── npm/
│   ├── gyuma-cli/                  # @goqoo/gyuma-cli パッケージ
│   │   └── package.json
│   ├── gyuma/                      # @goqoo/gyuma パッケージ（JSラッパー）
│   │   ├── package.json
│   │   ├── lib/
│   │   │   └── index.ts
│   │   └── bin/
│   │       └── gyuma.js
│   └── gyuma-platform/             # プラットフォームパッケージのテンプレート
│       └── package.json
├── go.mod
├── go.sum
├── docs/
│   └── go-rewrite-design.md
└── CLAUDE.md
```

### CLI インターフェース

```
gyuma [options]

  -d, --domain           kintone ドメイン名
  -i, --client-id        OAuth2 Client ID
  -s, --client-secret    OAuth2 Client Secret
  -S, --scope            OAuth2 Scope（スペース区切り or カンマ区切り）
  -P, --port             ローカルサーバーポート（デフォルト: 3000）
      --refresh-token    リフレッシュトークンを保存・利用する（デフォルト: 無効）
      --proxy            プロキシサーバー
      --pfx-filepath     クライアント証明書ファイルパス
      --pfx-password     クライアント証明書パスワード
      --noprompt         インタラクティブな入力を無効化
  -v, --version          バージョン表示
  -h, --help             ヘルプ表示
```

> 現行にあった `--password`（クレデンシャル暗号化パスワード）は暗号化廃止に伴い削除。

**標準出力**: `access_token` のみ（JSラッパーがキャプチャする）  
**標準エラー**: ログ・エラーメッセージ（ユーザーに見せる情報）

### OAuth フロー

```
1. tokens.json のキャッシュを確認
   - 存在しない → 新規取得へ
   - scope が違う → 新規取得へ
   - 有効期限切れ かつ --refresh-token 有効 → リフレッシュトークンで更新へ
   - 有効期限切れ かつ --refresh-token 無効 → 新規取得へ
   - 有効 → access_token を stdout に出力して終了

2. リフレッシュトークンによる更新フロー（--refresh-token 有効時）
   a. tokens.json から refresh_token を取得
   b. トークンエンドポイントに refresh_token グラントで POST
   c. 成功 → tokens.json を更新、access_token を stdout に出力して終了
   d. 失敗（期限切れ等）→ 新規取得フローへ

3. 新規取得フロー
   a. クレデンシャルを優先順位に従って取得（CLI → 環境変数 → ファイル → プロンプト）
   b. 自己署名証明書の確認・生成
   c. HTTPS サーバー起動（localhost:PORT）
   d. ブラウザ起動 → kintone 認可ページへリダイレクト
   e. コールバック受信（state 検証で CSRF 対策）
   f. トークンエンドポイントへ POST
   g. tokens.json を更新
   h. access_token を stdout に出力して終了
```

---

## @goqoo/gyuma（JSラッパー）の設計

### npm パッケージ構成

| パッケージ名 | 内容 |
|---|---|
| `@goqoo/gyuma` | JSラッパー（メインパッケージ） |
| `@goqoo/gyuma-darwin-arm64` | Goバイナリ（Apple Silicon） |
| `@goqoo/gyuma-darwin-x64` | Goバイナリ（Intel Mac） |
| `@goqoo/gyuma-linux-x64` | Goバイナリ（Linux x86_64） |
| `@goqoo/gyuma-linux-arm64` | Goバイナリ（Linux ARM64） |
| `@goqoo/gyuma-win32-x64` | Goバイナリ（Windows x64） |

### package.json

```json
{
  "name": "@goqoo/gyuma",
  "version": "1.0.0",
  "main": "./lib/index.js",
  "bin": {
    "gyuma": "./bin/gyuma.js"
  },
  "optionalDependencies": {
    "@goqoo/gyuma-darwin-arm64": "1.0.0",
    "@goqoo/gyuma-darwin-x64": "1.0.0",
    "@goqoo/gyuma-linux-x64": "1.0.0",
    "@goqoo/gyuma-linux-arm64": "1.0.0",
    "@goqoo/gyuma-win32-x64": "1.0.0"
  }
}
```

### JS API

```ts
import { gyuma } from '@goqoo/gyuma'

const token = await gyuma({
  domain: 'example.cybozu.com',
  client_id: process.env.GYUMA_CLIENT_ID,
  client_secret: process.env.GYUMA_CLIENT_SECRET,
  scope: 'k:app_settings:read k:app_settings:write',
})
```

### ラッパーの実装方針

```ts
import { spawn } from 'child_process'
import { resolveBinaryPath } from './resolve-binary'

export const gyuma = (argv: Argv): Promise<string> => {
  return new Promise((resolve, reject) => {
    const bin = resolveBinaryPath()
    const args = buildArgs(argv)

    const child = spawn(bin, args, {
      stdio: ['inherit', 'pipe', 'inherit'],
      //      ↑ stdin 継承      ↑ stdout キャプチャ  ↑ stderr 継承
    })

    let stdout = ''
    child.stdout.on('data', (data) => { stdout += data })
    child.on('close', (code) => {
      if (code === 0) resolve(stdout.trim())
      else reject(new Error(`gyuma exited with code ${code}`))
    })
  })
}
```

### `bin/gyuma.js` シム（CLI用）

Windows 対応のためシェルスクリプトではなく JS ファイルを採用。

```js
#!/usr/bin/env node
const { spawnSync } = require('child_process')
const { resolveBinaryPath } = require('../lib/resolve-binary')

const result = spawnSync(resolveBinaryPath(), process.argv.slice(2), {
  stdio: 'inherit',
})
process.exit(result.status ?? 1)
```

---

## リリース・配布フロー

### GitHub Actions によるクロスコンパイル

```yaml
# .github/workflows/release.yml（概略）
strategy:
  matrix:
    include:
      - goos: darwin,  goarch: arm64
      - goos: darwin,  goarch: amd64
      - goos: linux,   goarch: amd64
      - goos: linux,   goarch: arm64
      - goos: windows, goarch: amd64

steps:
  - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o gyuma ./cmd/gyuma
  - run: npm publish
```

---

## 移行計画

| フェーズ | 内容 | バージョン |
|---|---|---|
| **Phase 1** | `@goqoo/gyuma-cli` Goバイナリの実装・テスト | - |
| **Phase 2** | `@goqoo/gyuma-cli` Goバイナリの実装・テスト | - |
| **Phase 3** | `@goqoo/gyuma` JSラッパーの実装、npm パッケージ構成の整備 | 1.0.0-beta |
| **Phase 4** | 既存の `gyuma` npm パッケージを deprecated に（README・npm に明記） | 1.0.0 |
| **Phase 5** | 既存の `gyuma` npm パッケージのコードを削除 | 2.0.0（将来） |

Phase 4 では `package.json` の `deprecated` フィールドと README に `@goqoo/gyuma` への移行案内を記載する。  
Phase 5 は Phase 4 から最低 6 ヶ月以上の猶予を設ける。

### Node.js 版からの移行時の注意点

- クレデンシャルの暗号化ファイル（`~/.config/gyuma/<domain>/credentials`）は引き継ぎ不可。初回のみ再入力が必要。
- トークンキャッシュ（`~/.config/gyuma/<domain>/token.json`）もファイル形式が変わるため引き継ぎ不可。初回のみ再認証が必要。

---

## 検討事項・未決事項

- [ ] `--noprompt` モードでの CI/CD 利用シナリオを Go 版でも整備するか（将来検討）
- [ ] Windows での自己署名証明書の扱い（kintone は HTTPS 必須のため HTTP への変更は不可。開発者ツールとして自己署名証明書の警告を受け入れる運用を想定）
