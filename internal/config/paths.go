package config

import (
	"os"
	"path/filepath"
)

// ConfigDir は gyuma の設定ディレクトリパスを返す
// デフォルト: ~/.config/gyuma/
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gyuma"), nil
}

// TokensFile はトークンキャッシュファイルのパスを返す
// ~/.config/gyuma/tokens.json
func TokensFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tokens.json"), nil
}

// CredentialsFile はクレデンシャルファイルのパスを返す
// ~/.config/gyuma/credentials
func CredentialsFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials"), nil
}

// CertsDir は証明書ディレクトリのパスを返す
// ~/.config/gyuma/certs/
func CertsDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "certs"), nil
}

// MkcertCertFile は mkcert 証明書ファイルのパスを返す
// ~/.config/gyuma/certs/localhost.pem
func MkcertCertFile() (string, error) {
	dir, err := CertsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "localhost.pem"), nil
}

// MkcertKeyFile は mkcert 秘密鍵ファイルのパスを返す
// ~/.config/gyuma/certs/localhost-key.pem
func MkcertKeyFile() (string, error) {
	dir, err := CertsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "localhost-key.pem"), nil
}

// SelfSignedCertFile は自己署名証明書ファイルのパスを返す
// ~/.config/gyuma/certs/self-signed.pem
func SelfSignedCertFile() (string, error) {
	dir, err := CertsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "self-signed.pem"), nil
}

// SelfSignedKeyFile は自己署名秘密鍵ファイルのパスを返す
// ~/.config/gyuma/certs/self-signed-key.pem
func SelfSignedKeyFile() (string, error) {
	dir, err := CertsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "self-signed-key.pem"), nil
}

// EnsureConfigDir は設定ディレクトリが存在することを保証する
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

// EnsureCertsDir は証明書ディレクトリが存在することを保証する
func EnsureCertsDir() error {
	dir, err := CertsDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}
