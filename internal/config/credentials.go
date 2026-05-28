package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Credentials は OAuth クレデンシャルを保持する
type Credentials struct {
	ClientID     string
	ClientSecret string
}

// CredentialsSource はクレデンシャルの取得元を表す
type CredentialsSource int

const (
	SourceCLI CredentialsSource = iota
	SourceEnv
	SourceNetrcGPG
	SourceNetrc
	SourcePrompt
)

func (s CredentialsSource) String() string {
	switch s {
	case SourceCLI:
		return "CLI options"
	case SourceEnv:
		return "environment variables"
	case SourceNetrcGPG:
		return "~/.netrc.gpg"
	case SourceNetrc:
		return "~/.netrc"
	case SourcePrompt:
		return "interactive prompt"
	default:
		return "unknown"
	}
}

// GetCredentials は優先順位に従ってクレデンシャルを取得する
// 優先順位: CLI オプション → 環境変数 → ~/.netrc.gpg → ~/.netrc → プロンプト
func GetCredentials(domain, cliClientID, cliClientSecret string, noPrompt bool) (*Credentials, CredentialsSource, error) {
	// 1. CLI オプション
	if cliClientID != "" && cliClientSecret != "" {
		return &Credentials{
			ClientID:     cliClientID,
			ClientSecret: cliClientSecret,
		}, SourceCLI, nil
	}

	// 2. 環境変数
	envClientID := os.Getenv("GYUMA_CLIENT_ID")
	envClientSecret := os.Getenv("GYUMA_CLIENT_SECRET")
	if envClientID != "" && envClientSecret != "" {
		return &Credentials{
			ClientID:     envClientID,
			ClientSecret: envClientSecret,
		}, SourceEnv, nil
	}

	// 3. ~/.netrc.gpg（GPG 復号）
	if creds, err := loadFromNetrcGPG(domain); err == nil && creds != nil {
		return creds, SourceNetrcGPG, nil
	}

	// 4. ~/.netrc
	if creds, err := loadFromNetrc(domain); err == nil && creds != nil {
		return creds, SourceNetrc, nil
	}

	// 5. インタラクティブプロンプト
	if noPrompt {
		return nil, 0, fmt.Errorf("credentials not found and --noprompt is set")
	}

	creds, err := promptCredentials()
	if err != nil {
		return nil, 0, err
	}
	return creds, SourcePrompt, nil
}

// loadFromNetrc は ~/.netrc から <domain>:oauth クレデンシャルを読み込む
func loadFromNetrc(domain string) (*Credentials, error) {
	path, err := NetrcFile()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findOAuthCredentials(f, domain)
}

// loadFromNetrcGPG は ~/.netrc.gpg を gpg で復号して <domain>:oauth クレデンシャルを読み込む
func loadFromNetrcGPG(domain string) (*Credentials, error) {
	path, err := NetrcGPGFile()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	decrypted, err := decryptGPG(path)
	if err != nil {
		return nil, err
	}
	return findOAuthCredentials(bytes.NewReader(decrypted), domain)
}

// decryptGPG は gpg コマンドにファイルを渡して復号し平文を返す。
// パスフレーズ入力・鍵管理は gpg-agent / pinentry に委ねる（gyuma 自身は扱わない）。
func decryptGPG(path string) ([]byte, error) {
	cmd := exec.Command("gpg", "--quiet", "--decrypt", path)
	cmd.Stderr = os.Stderr // pinentry プロンプトや進捗メッセージはそのまま表示
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt %s with gpg: %w", path, err)
	}
	return out, nil
}

// promptCredentials はインタラクティブにクレデンシャルを入力させる
func promptCredentials() (*Credentials, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	clientID = strings.TrimSpace(clientID)

	fmt.Print("Client Secret: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	clientSecret = strings.TrimSpace(clientSecret)

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("client_id and client_secret are required")
	}

	return &Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
