# Gyuma

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

English | [日本語](/README.ja.md)

A CLI tool for kintone OAuth 2.0 authentication. Launches a local HTTPS server to handle the OAuth callback and outputs the access token.

## Installation

### npm

```bash
npm install -g gyuma
```

Or use npx without global installation:

```bash
npx gyuma -d example.cybozu.com -S "k:app_settings:read"
```

### Homebrew (macOS / Linux)

```bash
brew install goqoo-on-kintone/tap/gyuma
```

### go install

```bash
go install github.com/goqoo-on-kintone/gyuma/cmd/gyuma@latest
```

### Binary Download

Download platform-specific binaries from [GitHub Releases](https://github.com/goqoo-on-kintone/gyuma/releases):

- `gyuma_X.X.X_darwin_amd64.tar.gz` (macOS Intel)
- `gyuma_X.X.X_darwin_arm64.tar.gz` (macOS Apple Silicon)
- `gyuma_X.X.X_linux_amd64.tar.gz` (Linux x64)
- `gyuma_X.X.X_linux_arm64.tar.gz` (Linux ARM64)
- `gyuma_X.X.X_windows_amd64.zip` (Windows x64)
- `gyuma_X.X.X_windows_arm64.zip` (Windows ARM64)

### Build from Source

```bash
git clone https://github.com/goqoo-on-kintone/gyuma.git
cd gyuma
make build
# Binary is generated at bin/gyuma
```

## Usage

### Get Access Token

```bash
# Basic usage (interactive mode)
gyuma -d example.cybozu.com -S "k:app_settings:read"

# With credentials
gyuma -d example.cybozu.com \
  -i YOUR_CLIENT_ID \
  -s YOUR_CLIENT_SECRET \
  -S "k:app_settings:read k:app_settings:write"

# Use with refresh token
gyuma -d example.cybozu.com -S "k:app_settings:read" --refresh-token

# Save credentials for future use
gyuma -d example.cybozu.com -i YOUR_CLIENT_ID -s YOUR_CLIENT_SECRET \
  -S "k:app_settings:read" --save-credentials
```

### Setup Certificate

For smoother OAuth flow without browser SSL warnings, use mkcert certificates:

```bash
# Install mkcert first (https://github.com/FiloSottile/mkcert)
brew install mkcert
mkcert -install

# Setup certificate for gyuma
gyuma setup-cert
```

## Options

### Main Command

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--domain` | `-d` | | kintone domain name (required) |
| `--client-id` | `-i` | | OAuth2 Client ID |
| `--client-secret` | `-s` | | OAuth2 Client Secret |
| `--scope` | `-S` | | OAuth2 Scope (required) |
| `--port` | `-P` | `3000` | Local server port |
| `--refresh-token` | | `false` | Save and use refresh token |
| `--save-credentials` | | `false` | Save credentials to file |
| `--noprompt` | | `false` | Disable interactive input |
| `--quiet` | | `false` | Suppress warning messages |
| `--version` | `-v` | | Show version |
| `--help` | `-h` | | Show help |

### setup-cert Subcommand

| Option | Default | Description |
|--------|---------|-------------|
| `--host` | `localhost` | Certificate hostname |
| `--port` | `3000` | HTTPS server port |
| `--force` | `false` | Overwrite existing certificate |

## Configuration

### Credentials Priority

Credentials are resolved in the following order:

1. CLI options (`-i`, `-s`)
2. Environment variables (`GYUMA_CLIENT_ID`, `GYUMA_CLIENT_SECRET`)
3. Credentials file (`~/.config/gyuma/credentials`)
4. Interactive prompt (if not `--noprompt`)

### File Locations

| File | Path | Description |
|------|------|-------------|
| Tokens | `~/.config/gyuma/tokens.json` | Cached access/refresh tokens |
| Credentials | `~/.config/gyuma/credentials` | Saved client credentials (INI format) |
| Certificates | `~/.config/gyuma/certs/` | SSL certificates for local server |

### Environment Variables

```bash
export GYUMA_CLIENT_ID="your_client_id"
export GYUMA_CLIENT_SECRET="your_client_secret"
```

## Use with Other Tools

### Shell Script

```bash
TOKEN=$(gyuma -d example.cybozu.com -S "k:app_settings:read")

curl -H "Authorization: Bearer $TOKEN" \
  "https://example.cybozu.com/k/v1/app/form/fields.json?app=1"
```

### ginue

[ginue](https://github.com/goqoo-on-kintone/ginue) can use gyuma for OAuth authentication:

```bash
ginue pull --oauth
```

## kintone OAuth Documentation

- [How to add OAuth clients - English](https://kintone.dev/en/docs/common/authentication/how-to-add-oauth-clients/)
- [OAuthクライアントの使用 - 日本語](https://cybozu.dev/ja/kintone/docs/common/authentication/how-to-add-oauth-clients/)

## Development

```bash
# Build
make build

# Test
make test

# Cross-compile for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.
