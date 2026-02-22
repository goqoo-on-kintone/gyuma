package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsSource_String(t *testing.T) {
	tests := []struct {
		source CredentialsSource
		want   string
	}{
		{SourceCLI, "CLI options"},
		{SourceEnv, "environment variables"},
		{SourceFile, "credentials file"},
		{SourcePrompt, "interactive prompt"},
		{CredentialsSource(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.source.String(); got != tt.want {
				t.Errorf("CredentialsSource.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCredentials_CLI(t *testing.T) {
	creds, source, err := GetCredentials("example.cybozu.com", "cli-id", "cli-secret", true)
	if err != nil {
		t.Fatalf("GetCredentials() returned error: %v", err)
	}
	if source != SourceCLI {
		t.Errorf("source = %v, want SourceCLI", source)
	}
	if creds.ClientID != "cli-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "cli-id")
	}
	if creds.ClientSecret != "cli-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "cli-secret")
	}
}

func TestGetCredentials_Env(t *testing.T) {
	// 環境変数を設定
	oldID := os.Getenv("GYUMA_CLIENT_ID")
	oldSecret := os.Getenv("GYUMA_CLIENT_SECRET")
	defer func() {
		os.Setenv("GYUMA_CLIENT_ID", oldID)
		os.Setenv("GYUMA_CLIENT_SECRET", oldSecret)
	}()

	os.Setenv("GYUMA_CLIENT_ID", "env-id")
	os.Setenv("GYUMA_CLIENT_SECRET", "env-secret")

	creds, source, err := GetCredentials("example.cybozu.com", "", "", true)
	if err != nil {
		t.Fatalf("GetCredentials() returned error: %v", err)
	}
	if source != SourceEnv {
		t.Errorf("source = %v, want SourceEnv", source)
	}
	if creds.ClientID != "env-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "env-id")
	}
	if creds.ClientSecret != "env-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "env-secret")
	}
}

func TestGetCredentials_NoPromptError(t *testing.T) {
	// 環境変数をクリア
	oldID := os.Getenv("GYUMA_CLIENT_ID")
	oldSecret := os.Getenv("GYUMA_CLIENT_SECRET")
	defer func() {
		os.Setenv("GYUMA_CLIENT_ID", oldID)
		os.Setenv("GYUMA_CLIENT_SECRET", oldSecret)
	}()

	os.Unsetenv("GYUMA_CLIENT_ID")
	os.Unsetenv("GYUMA_CLIENT_SECRET")

	// noPrompt=true でクレデンシャルが見つからない場合はエラー
	_, _, err := GetCredentials("nonexistent.cybozu.com", "", "", true)
	if err == nil {
		t.Error("GetCredentials() should return error when noPrompt=true and no credentials")
	}
}

func TestSaveAndLoadCredentials(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "gyuma-creds-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// ConfigDir を一時的にオーバーライドできないため、実際のファイルを使用
	// このテストは ~/.config/gyuma/credentials に影響を与える可能性があるためスキップ
	t.Skip("This test modifies real credentials file - run manually if needed")
}

func TestLoadCredentialsFromFile(t *testing.T) {
	// テスト用のクレデンシャルファイルを作成
	tmpDir, err := os.MkdirTemp("", "gyuma-creds-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 一時的なクレデンシャルファイルを作成
	credPath := filepath.Join(tmpDir, "credentials")
	content := `[example.cybozu.com]
client_id = test-client-id
client_secret = test-client-secret

[another.cybozu.com]
client_id = another-id
client_secret = another-secret
`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test credentials file: %v", err)
	}

	// 注意: loadCredentialsFromFile は CredentialsFile() を使用するため、
	// 実際のファイルパスをオーバーライドできない。
	// この関数のテストは統合テストで行う必要がある。
	t.Skip("loadCredentialsFromFile uses fixed path - integration test required")
}
