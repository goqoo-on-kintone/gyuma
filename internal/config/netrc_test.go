package config

import (
	"strings"
	"testing"
)

func TestFindOAuthCredentials(t *testing.T) {
	// 同一ホストに kintone ユーザー認証と :oauth が併存していても :oauth を引く
	const content = `machine example.cybozu.com
  login kintone-user
  password kintone-pass

machine example.cybozu.com:oauth
  login my-client-id
  password my-client-secret
`
	creds, err := findOAuthCredentials(strings.NewReader(content), "example.cybozu.com")
	if err != nil {
		t.Fatalf("findOAuthCredentials() returned error: %v", err)
	}
	if creds.ClientID != "my-client-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "my-client-id")
	}
	if creds.ClientSecret != "my-client-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "my-client-secret")
	}
}

func TestFindOAuthCredentials_NotFound(t *testing.T) {
	// :oauth エントリが無い場合は、ユーザー認証の machine を誤って拾わずエラー
	const content = `machine example.cybozu.com
  login kintone-user
  password kintone-pass
`
	if _, err := findOAuthCredentials(strings.NewReader(content), "example.cybozu.com"); err == nil {
		t.Fatal("findOAuthCredentials() should return error when no :oauth entry exists")
	}
}

func TestFindOAuthCredentials_IgnoresDefault(t *testing.T) {
	// default エントリがあっても :oauth が無ければフォールバックしない
	const content = `default
  login default-user
  password default-pass
`
	if _, err := findOAuthCredentials(strings.NewReader(content), "example.cybozu.com"); err == nil {
		t.Fatal("findOAuthCredentials() should not fall back to default machine")
	}
}
