# Go 全面書き直し 設計書

## 背景・目的

現在の gyuma は Node.js / TypeScript 製のライブラリ・CLI ツールであり、利用するには Node.js ランタイムが必要。  
Go で書き直すことで以下を実現する：

- **単一バイナリ配布**: Node.js 不要。どの言語・環境からでも使いやすい
- **クロスコンパイル**: `GOOS` / `GOARCH` の指定だけで全プラットフォーム向けバイナリを生成
- **Node.js ライブラリとしての後方互換維持**: 既存ユーザーへの影響をゼロにする

---

## 採用アーキテクチャ：Goバイナリ ＋ 薄いNode.jsラッパー

esbuild・Biome・Prisma・SWC など、Go/Rust 製ツールを npm でも配布している主要 OSS が採用しているパターンを踏襲する。

```
┌─────────────────────────────────────────┐
│  呼び出し元                              │
│  ・JS/TS から require('gyuma')           │
│  ・シェルスクリプトから gyuma CLI         │
│  ・Python/Ruby など他言語から バイナリ直接 │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│  Node.js ラッパー（薄い）                │
│  gyuma npm パッケージ                    │
│  ・JS API (require/import) を提供        │
│  ・内部で Go バイナリを child_process で起動│
│  ・access_token を stdout からキャプチャ  │
└────────────┬────────────────────────────┘
             │ execFile / spawn
┌────────────▼────────────────────────────┐
│  Go バイナリ                             │
│  ・OAuth フロー全体を担う                 │
│  ・HTTPS サーバー起動                    │
│  ・ブラウザ自動起動                       │
│  ・トークンキャッシュ（ファイル）          │
│  ・access_token を stdout に出力して終了  │
└─────────────────────────────────────────┘
```

---

## npm パッケージ構成

メインパッケージに加え、プラットフォームごとのバイナリパッケージを optionalDependencies として配布する。

| パッケージ名 | 内容 |
|---|---|
| `gyuma` | JS ラッパー（メインパッケージ） |
| `gyuma-darwin-arm64` | Go バイナリ（Apple Silicon） |
| `gyuma-darwin-x64` | Go バイナリ（Intel Mac） |
| `gyuma-linux-x64` | Go バイナリ（Linux x86_64） |
| `gyuma-linux-arm64` | Go バイナリ（Linux ARM64） |
| `gyuma-win32-x64` | Go バイナリ（Windows x64） |

### メインパッケージの package.json

```json
{
  "name": "gyuma",
  "version": "1.0.0",
  "main": "./lib/index.js",
  "bin": {
    "gyuma": "./bin/gyuma"
  },
  "optionalDependencies": {
    "gyuma-darwin-arm64": "1.0.0",
    "gyuma-darwin-x64": "1.0.0",
    "gyuma-linux-x64": "1.0.0",
    "gyuma-linux-arm64": "1.0.0",
    "gyuma-win32-x64": "1.0.0"
  }
}
```

`npm install` 時に npm が `os` / `cpu` フィールドを見て、該当プラットフォームのバイナリパッケージだけをインストールする。

---

## Node.js ラッパー（JS API）の設計

### 後方互換 API

既存ユーザーのコードをそのまま動かす。インターフェースは現行の TypeScript 版と同一。

```ts
// 呼び出し側は今と変わらない
import { gyuma } from 'gyuma'

const token = await gyuma({
  domain: 'example.cybozu.com',
  client_id: process.env.OAUTH2_CLIENT_ID,
  client_secret: process.env.OAUTH2_CLIENT_SECRET,
  scope: 'k:app_settings:read k:app_settings:write',
  password: 'xxxxx',
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
      //        （パスワード入力など）                   （エラー表示）
    })

    let stdout = ''
    child.stdout.on('data', (data) => { stdout += data })
    child.on('close', (code) => {
      if (code === 0) resolve(stdout.trim()) // access_token
      else reject(new Error(`gyuma exited with code ${code}`))
    })
  })
}
```

**stdin を `inherit` にする理由**: gyuma はブラウザ起動・パスワード入力などインタラクティブな処理を含む。  
**stdout のみ `pipe` にする理由**: access_token だけをキャプチャしたい。エラーメッセージは stderr 経由でターミナルに直接出す。

### バイナリパス解決

```ts
// resolve-binary.ts
import { platform, arch } from 'os'
import { join } from 'path'

export const resolveBinaryPath = (): string => {
  const os = platform()   // darwin / linux / win32
  const cpu = arch()      // arm64 / x64

  const pkgName = `gyuma-${os}-${cpu}`

  try {
    // optionalDependencies からバイナリパスを取得
    return require.resolve(`${pkgName}/bin/gyuma`)
  } catch {
    throw new Error(
      `Unsupported platform: ${os}/${cpu}. ` +
      `Please install the appropriate package manually: npm install ${pkgName}`
    )
  }
}
```

---

## Go バイナリの設計

### ディレクトリ構成（Goプロジェクト）

```
gyuma/                        ← リポジトリルート
├── cmd/
│   └── gyuma/
│       └── main.go           # エントリポイント
├── internal/
│   ├── auth/
│   │   ├── server.go         # HTTPS サーバー / OAuth フロー
│   │   └── token.go          # トークンキャッシュ読み書き
│   ├── config/
│   │   ├── credentials.go    # クレデンシャル暗号化・保存
│   │   └── paths.go          # 設定ディレクトリパス
│   ├── browser/
│   │   └── open.go           # ブラウザ起動（OS別）
│   └── cert/
│       └── cert.go           # 自己署名証明書生成
├── npm/                      # npm 関連ファイル
│   ├── gyuma/                # メインパッケージ
│   │   ├── package.json
│   │   ├── lib/
│   │   │   └── index.ts      # JS ラッパー
│   │   └── bin/
│   │       └── gyuma         # CLI シム（バイナリを呼ぶだけのシェル or JS）
│   └── gyuma-platform/       # プラットフォームパッケージのテンプレート
│       └── package.json
├── go.mod
├── go.sum
├── docs/
│   └── go-rewrite-design.md  # 本ドキュメント
└── CLAUDE.md
```

### Go バイナリの CLI インターフェース

現行の Node.js CLI と同一のオプションを提供する（後方互換）。

```
gyuma [options]

  -d, --domain         kintone ドメイン名
  -i, --client-id      OAuth2 Client ID
  -s, --client-secret  OAuth2 Client Secret
  -S, --scope          OAuth2 Scope（スペース区切り or カンマ区切り）
  -p, --password       クレデンシャル暗号化パスワード
  -P, --port           ローカルサーバーポート（デフォルト: 3000）
      --proxy          プロキシサーバー
      --pfx-filepath   クライアント証明書ファイルパス
      --pfx-password   クライアント証明書パスワード
      --noprompt       インタラクティブな入力を無効化
  -v, --version        バージョン表示
  -h, --help           ヘルプ表示
```

**標準出力**: `access_token` のみ（JS ラッパーがキャプチャする）  
**標準エラー**: ログ・エラーメッセージ（ユーザーに見せる情報）

### OAuth フロー（Go実装）

現行の TypeScript 版と同等のロジックを Go で実装する。

```
1. キャッシュ済みトークンを確認
   - 存在しない → 新規取得へ
   - scope が違う → 新規取得へ
   - 有効期限切れ → 新規取得へ
   - 有効 → access_token を stdout に出力して終了

2. 新規取得フロー
   a. パスワード入力（--noprompt でなければ）
   b. クレデンシャル読み込み or 入力
   c. 自己署名証明書の確認・生成
   d. HTTPS サーバー起動（localhost:PORT）
   e. ブラウザ起動 → kintone 認可ページへリダイレクト
   f. コールバック受信（state 検証で CSRF 対策）
   g. トークンエンドポイントへ POST
   h. トークンをファイルに保存、クレデンシャルを暗号化保存
   i. access_token を stdout に出力して終了
```

### ファイル保存先（現行と同一）

- トークン: `~/.config/gyuma/<domain>/token.json`
- クレデンシャル: `~/.config/gyuma/<domain>/credentials`（AES暗号化）

> Go 版でも同じパスを使うことで、Node.js 版からの移行時にトークンキャッシュが引き継がれる。

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
  - run: npm publish  # 各プラットフォームパッケージを publish
```

Go のクロスコンパイルは `GOOS` と `GOARCH` を指定するだけで CGO 不使用なら追加ツール不要。

---

## 移行計画

| フェーズ | 内容 | バージョン |
|---|---|---|
| **Phase 1** | Go バイナリの実装・テスト | - |
| **Phase 2** | npm パッケージ構成の整備・JS ラッパー実装 | 1.0.0-beta |
| **Phase 3** | Node.js 版を deprecated に（README・npm に明記） | 1.0.0 |
| **Phase 4** | Node.js 版コードをリポジトリから削除 | 2.0.0（将来） |

Phase 3 では `package.json` の `deprecated` フィールドと README に移行案内を記載する。  
Phase 4 は Phase 3 から最低 6 ヶ月以上の猶予を設ける。

---

## クレデンシャルの暗号化設計

### 現行（Node.js版）の問題点

現行の `encrypt.ts` には以下のセキュリティ上の問題がある：

1. **ソルトが固定値**（致命的）
   ```ts
   const SALT = 'lKzR+i6IwG/DbuAY5thksw=='  // 全ユーザー共通
   ```
   ソルトが固定だとレインボーテーブル攻撃が有効になり、同じパスワードを使うユーザー間で同じ鍵が生成される。

2. **AES-256-CBC モード**（軽微）
   改ざん検知ができない。ファイルを書き換えられても気づけない。

### Go版での採用方式

| 項目 | 現行（Node.js） | Go 版 |
|---|---|---|
| KDF | scrypt | scrypt（継続） |
| ソルト | **固定値** | **ランダム生成・ファイルに同梱** |
| 暗号化モード | AES-256-CBC | **AES-256-GCM** |
| 改ざん検知 | なし | **あり**（GCM の AuthTag） |

互換性は維持しない。Go 版インストール後の初回実行時に再認証が必要になる。

### ファイルフォーマット（Go版）

```
[salt 32byte][nonce 12byte][ciphertext][authTag 16byte]
→ Base64 エンコードして保存
```

### scrypt パラメータ

対話型用途の標準的な値を採用する：

```go
key, _ := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
// N=32768（CPU/メモリコスト）, r=8, p=1
```

---

## `bin/gyuma` シムの設計

Windows 対応のため、シェルスクリプトではなく **JS ファイル** を採用する。

```js
#!/usr/bin/env node
// bin/gyuma.js
const { spawnSync } = require('child_process')
const { resolveBinaryPath } = require('../lib/resolve-binary')

const result = spawnSync(resolveBinaryPath(), process.argv.slice(2), {
  stdio: 'inherit',  // stdin/stdout/stderr すべて継承（CLIとして透過的に動作）
})
process.exit(result.status ?? 1)
```

CLI として使う場合は stdio をすべて `inherit` にする（JS API と異なる点）。

---

## 検討事項・未決事項

- [ ] `refresh_token` の扱い：現行は保存しない設計だが、Go 版で対応するか
- [ ] `noprompt` モードでの CI/CD 利用シナリオを Go 版でも整備するか
- [ ] Windows での自己署名証明書の扱い（trust store への登録が必要な場合がある）
