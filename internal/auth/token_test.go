package auth

import (
	"testing"
	"time"
)

func TestToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  bool
	}{
		{
			name: "有効なトークン（十分な残り時間）",
			token: Token{
				AccessToken: "valid_token",
				Expiry:      time.Now().Add(10 * time.Minute).Format(time.RFC3339),
			},
			want: true,
		},
		{
			name: "期限切れのトークン",
			token: Token{
				AccessToken: "expired_token",
				Expiry:      time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			},
			want: false,
		},
		{
			name: "残り1分未満のトークン",
			token: Token{
				AccessToken: "almost_expired_token",
				Expiry:      time.Now().Add(30 * time.Second).Format(time.RFC3339),
			},
			want: false,
		},
		{
			name: "空のアクセストークン",
			token: Token{
				AccessToken: "",
				Expiry:      time.Now().Add(10 * time.Minute).Format(time.RFC3339),
			},
			want: false,
		},
		{
			name: "不正な有効期限フォーマット",
			token: Token{
				AccessToken: "valid_token",
				Expiry:      "invalid-date-format",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("Token.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_HasScope(t *testing.T) {
	tests := []struct {
		name          string
		tokenScope    string
		requiredScope string
		want          bool
	}{
		{
			name:          "単一スコープ - 一致",
			tokenScope:    "k:app_record:read",
			requiredScope: "k:app_record:read",
			want:          true,
		},
		{
			name:          "単一スコープ - 不一致",
			tokenScope:    "k:app_record:read",
			requiredScope: "k:app_record:write",
			want:          false,
		},
		{
			name:          "複数スコープ - 全て含む",
			tokenScope:    "k:app_record:read k:app_record:write k:app_settings:read",
			requiredScope: "k:app_record:read k:app_settings:read",
			want:          true,
		},
		{
			name:          "複数スコープ - 一部不足",
			tokenScope:    "k:app_record:read",
			requiredScope: "k:app_record:read k:app_record:write",
			want:          false,
		},
		{
			name:          "空の要求スコープ",
			tokenScope:    "k:app_record:read",
			requiredScope: "",
			want:          true,
		},
		{
			name:          "空のトークンスコープ",
			tokenScope:    "",
			requiredScope: "k:app_record:read",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{Scope: tt.tokenScope}
			if got := token.HasScope(tt.requiredScope); got != tt.want {
				t.Errorf("Token.HasScope(%q) = %v, want %v", tt.requiredScope, got, tt.want)
			}
		})
	}
}
