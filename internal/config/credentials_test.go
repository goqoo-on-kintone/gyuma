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
		{SourceNetrcGPG, "~/.netrc.gpg"},
		{SourceNetrc, "~/.netrc"},
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

func TestGetCredentials_Netrc(t *testing.T) {
	// CLI・環境変数が無いとき ~/.netrc の <domain>:oauth から取得できる
	t.Setenv("GYUMA_CLIENT_ID", "")
	t.Setenv("GYUMA_CLIENT_SECRET", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	netrc := `machine example.cybozu.com
  login kintone-user
  password kintone-pass

machine example.cybozu.com:oauth
  login netrc-client-id
  password netrc-client-secret
`
	if err := os.WriteFile(filepath.Join(home, ".netrc"), []byte(netrc), 0600); err != nil {
		t.Fatalf("failed to write .netrc: %v", err)
	}

	creds, source, err := GetCredentials("example.cybozu.com", "", "", true)
	if err != nil {
		t.Fatalf("GetCredentials() returned error: %v", err)
	}
	if source != SourceNetrc {
		t.Errorf("source = %v, want SourceNetrc", source)
	}
	if creds.ClientID != "netrc-client-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "netrc-client-id")
	}
	if creds.ClientSecret != "netrc-client-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "netrc-client-secret")
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

	// 実ホームの ~/.netrc / ~/.netrc.gpg を参照して gpg を起動しないよう HOME を隔離する
	t.Setenv("HOME", t.TempDir())

	// noPrompt=true でクレデンシャルが見つからない場合はエラー
	_, _, err := GetCredentials("nonexistent.cybozu.com", "", "", true)
	if err == nil {
		t.Error("GetCredentials() should return error when noPrompt=true and no credentials")
	}
}
