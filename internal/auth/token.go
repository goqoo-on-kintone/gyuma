package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goqoo-on-kintone/gyuma/internal/config"
)

// Token は OAuth トークン情報を保持する
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	Expiry       string `json:"expiry"`
}

// TokenStore はドメインごとのトークンを保持する
type TokenStore map[string]*Token

// IsValid はトークンが有効かどうかを判定する
func (t *Token) IsValid() bool {
	if t.AccessToken == "" {
		return false
	}

	expiry, err := time.Parse(time.RFC3339, t.Expiry)
	if err != nil {
		return false
	}

	// 有効期限まで1分以上あれば有効とみなす
	return time.Until(expiry) > time.Minute
}

// HasScope は指定されたスコープが全て含まれているかを判定する
func (t *Token) HasScope(requiredScope string) bool {
	tokenScopes := strings.Fields(t.Scope)
	requiredScopes := strings.Fields(requiredScope)

	tokenScopeSet := make(map[string]bool)
	for _, s := range tokenScopes {
		tokenScopeSet[s] = true
	}

	for _, s := range requiredScopes {
		if !tokenScopeSet[s] {
			return false
		}
	}
	return true
}

// LoadTokenStore はトークンストアをファイルから読み込む
func LoadTokenStore() (TokenStore, error) {
	path, err := config.TokensFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(TokenStore), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	var store TokenStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse tokens file: %w", err)
	}

	return store, nil
}

// SaveTokenStore はトークンストアをファイルに保存する
func SaveTokenStore(store TokenStore) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	path, err := config.TokensFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	return nil
}

// GetToken は指定されたドメインのトークンを取得する
func GetToken(domain string) (*Token, error) {
	store, err := LoadTokenStore()
	if err != nil {
		return nil, err
	}

	token, ok := store[domain]
	if !ok {
		return nil, nil
	}

	return token, nil
}

// SaveToken は指定されたドメインのトークンを保存する
func SaveToken(domain string, token *Token) error {
	store, err := LoadTokenStore()
	if err != nil {
		return err
	}

	store[domain] = token
	return SaveTokenStore(store)
}

// GetValidToken は有効なトークンを取得する
// トークンが存在しない、期限切れ、またはスコープが不足している場合は nil を返す
func GetValidToken(domain, requiredScope string) (*Token, error) {
	token, err := GetToken(domain)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, nil
	}

	if !token.IsValid() {
		return nil, nil
	}

	if !token.HasScope(requiredScope) {
		return nil, nil
	}

	return token, nil
}
