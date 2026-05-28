package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
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
	SourceFile
	SourcePrompt
)

func (s CredentialsSource) String() string {
	switch s {
	case SourceCLI:
		return "CLI options"
	case SourceEnv:
		return "environment variables"
	case SourceFile:
		return "credentials file"
	case SourcePrompt:
		return "interactive prompt"
	default:
		return "unknown"
	}
}

// GetCredentials は優先順位に従ってクレデンシャルを取得する
// 優先順位: CLI オプション → 環境変数 → ファイル → プロンプト
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

	// 3. クレデンシャルファイル
	creds, err := loadCredentialsFromFile(domain)
	if err == nil && creds != nil {
		return creds, SourceFile, nil
	}

	// 4. インタラクティブプロンプト
	if noPrompt {
		return nil, 0, fmt.Errorf("credentials not found and --noprompt is set")
	}

	creds, err = promptCredentials()
	if err != nil {
		return nil, 0, err
	}
	return creds, SourcePrompt, nil
}

// loadCredentialsFromFile はクレデンシャルファイルから指定ドメインのクレデンシャルを読み込む
func loadCredentialsFromFile(domain string) (*Credentials, error) {
	path, err := CredentialsFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found: %s", path)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials file: %w", err)
	}

	section, err := cfg.GetSection(domain)
	if err != nil {
		return nil, fmt.Errorf("domain section not found: %s", domain)
	}

	clientID := section.Key("client_id").String()
	clientSecret := section.Key("client_secret").String()

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("incomplete credentials for domain: %s", domain)
	}

	return &Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
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

// SaveCredentials はクレデンシャルをファイルに保存する
func SaveCredentials(domain string, creds *Credentials) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	path, err := CredentialsFile()
	if err != nil {
		return err
	}

	var cfg *ini.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = ini.Empty()
	} else {
		cfg, err = ini.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load credentials file: %w", err)
		}
	}

	section, err := cfg.NewSection(domain)
	if err != nil {
		return fmt.Errorf("failed to create section: %w", err)
	}

	section.Key("client_id").SetValue(creds.ClientID)
	section.Key("client_secret").SetValue(creds.ClientSecret)

	if err := cfg.SaveTo(path); err != nil {
		return fmt.Errorf("failed to save credentials file: %w", err)
	}

	// パーミッションを 600 に設定
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
