package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() returned error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "gyuma")
	if dir != expected {
		t.Errorf("ConfigDir() = %q, want %q", dir, expected)
	}
}

func TestTokensFile(t *testing.T) {
	path, err := TokensFile()
	if err != nil {
		t.Fatalf("TokensFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("gyuma", "tokens.json")) {
		t.Errorf("TokensFile() = %q, want suffix 'gyuma/tokens.json'", path)
	}
}

func TestCredentialsFile(t *testing.T) {
	path, err := CredentialsFile()
	if err != nil {
		t.Fatalf("CredentialsFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("gyuma", "credentials")) {
		t.Errorf("CredentialsFile() = %q, want suffix 'gyuma/credentials'", path)
	}
}

func TestCertsDir(t *testing.T) {
	dir, err := CertsDir()
	if err != nil {
		t.Fatalf("CertsDir() returned error: %v", err)
	}

	if !strings.HasSuffix(dir, filepath.Join("gyuma", "certs")) {
		t.Errorf("CertsDir() = %q, want suffix 'gyuma/certs'", dir)
	}
}

func TestMkcertCertFile(t *testing.T) {
	path, err := MkcertCertFile()
	if err != nil {
		t.Fatalf("MkcertCertFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("certs", "localhost.pem")) {
		t.Errorf("MkcertCertFile() = %q, want suffix 'certs/localhost.pem'", path)
	}
}

func TestMkcertKeyFile(t *testing.T) {
	path, err := MkcertKeyFile()
	if err != nil {
		t.Fatalf("MkcertKeyFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("certs", "localhost-key.pem")) {
		t.Errorf("MkcertKeyFile() = %q, want suffix 'certs/localhost-key.pem'", path)
	}
}

func TestSelfSignedCertFile(t *testing.T) {
	path, err := SelfSignedCertFile()
	if err != nil {
		t.Fatalf("SelfSignedCertFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("certs", "self-signed.pem")) {
		t.Errorf("SelfSignedCertFile() = %q, want suffix 'certs/self-signed.pem'", path)
	}
}

func TestSelfSignedKeyFile(t *testing.T) {
	path, err := SelfSignedKeyFile()
	if err != nil {
		t.Fatalf("SelfSignedKeyFile() returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("certs", "self-signed-key.pem")) {
		t.Errorf("SelfSignedKeyFile() = %q, want suffix 'certs/self-signed-key.pem'", path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// EnsureConfigDir を呼び出し
	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() returned error: %v", err)
	}

	// ディレクトリが存在することを確認
	dir, _ := ConfigDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Config directory does not exist after EnsureConfigDir(): %v", err)
	}
	if !info.IsDir() {
		t.Error("ConfigDir path exists but is not a directory")
	}
}

func TestEnsureCertsDir(t *testing.T) {
	// EnsureCertsDir を呼び出し
	err := EnsureCertsDir()
	if err != nil {
		t.Fatalf("EnsureCertsDir() returned error: %v", err)
	}

	// ディレクトリが存在することを確認
	dir, _ := CertsDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Certs directory does not exist after EnsureCertsDir(): %v", err)
	}
	if !info.IsDir() {
		t.Error("CertsDir path exists but is not a directory")
	}
}
