package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/goqoo-on-kintone/gyuma/internal/cert"
)

// OAuthConfig は OAuth 認証の設定を保持する
type OAuthConfig struct {
	Domain       string
	ClientID     string
	ClientSecret string
	Scope        string
	Port         int
	UseRefresh   bool
	Quiet        bool
}

// tokenResponse は kintone トークンエンドポイントのレスポンス
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// RunOAuthFlow は OAuth 認証フローを実行し、アクセストークンを返す
func RunOAuthFlow(cfg *OAuthConfig) (string, error) {
	// キャッシュされたトークンを確認
	token, err := GetValidToken(cfg.Domain, cfg.Scope)
	if err != nil {
		return "", err
	}
	if token != nil {
		return token.AccessToken, nil
	}

	// リフレッシュトークンがある場合は更新を試みる
	if cfg.UseRefresh {
		existingToken, _ := GetToken(cfg.Domain)
		if existingToken != nil && existingToken.RefreshToken != "" {
			newToken, err := refreshAccessToken(cfg, existingToken.RefreshToken)
			if err == nil {
				return newToken, nil
			}
			// リフレッシュ失敗時は通常フローにフォールバック
			fmt.Fprintln(os.Stderr, "Refresh token expired, starting new authorization...")
		}
	}

	// 証明書を取得
	certPaths, isMkcert, err := cert.GetCertPaths()
	if err != nil {
		return "", fmt.Errorf("failed to get certificate: %w", err)
	}

	// mkcert 証明書がない場合は警告を表示
	if !isMkcert && !cfg.Quiet {
		fmt.Fprintln(os.Stderr, "⚠ mkcert証明書が見つかりません。ブラウザで警告が表示される場合があります。")
		fmt.Fprintln(os.Stderr, "  Tip: `gyuma setup-cert` で信頼された証明書をセットアップできます。")
		fmt.Fprintln(os.Stderr, "")
	}

	// state パラメータを生成（CSRF 対策）
	state, err := generateState()
	if err != nil {
		return "", err
	}

	// コールバック用のチャネル
	resultCh := make(chan string, 1)
	errorCh := make(chan error, 1)

	// HTTPS サーバーを起動
	redirectURI := fmt.Sprintf("https://localhost:%d/oauth2callback", cfg.Port)
	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port)}

	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		// state を検証
		if r.URL.Query().Get("state") != state {
			errorCh <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// 認可コードを取得
		code := r.URL.Query().Get("code")
		if code == "" {
			errorMsg := r.URL.Query().Get("error")
			errorDesc := r.URL.Query().Get("error_description")
			errorCh <- fmt.Errorf("authorization error: %s - %s", errorMsg, errorDesc)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		// トークンを取得
		token, err := exchangeCode(cfg, code, redirectURI)
		if err != nil {
			errorCh <- err
			http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
			return
		}

		// 成功ページを表示
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>Authorization Successful</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; text-align: center; padding: 50px; }
    h1 { color: #333; }
    p { color: #666; font-size: 18px; }
  </style>
</head>
<body>
<h1>🎉 Authorization Successful!</h1>
<p>✨ You can close this window now ✨</p>
<script>window.close();</script>
</body>
</html>`)

		resultCh <- token
	})

	// サーバーを goroutine で起動
	go func() {
		if err := server.ListenAndServeTLS(certPaths.CertFile, certPaths.KeyFile); err != http.ErrServerClosed {
			errorCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// 認可 URL を生成してブラウザを開く
	authURL := buildAuthURL(cfg, state, redirectURI)
	fmt.Fprintf(os.Stderr, "Opening browser for authorization...\n")
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser. Please visit:\n%s\n\n", authURL)
	}

	// 結果を待機
	select {
	case accessToken := <-resultCh:
		server.Shutdown(context.Background())
		return accessToken, nil
	case err := <-errorCh:
		server.Shutdown(context.Background())
		return "", err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return "", fmt.Errorf("authorization timeout")
	}
}

// generateState は CSRF 対策用の state パラメータを生成する
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// buildAuthURL は kintone の認可 URL を生成する
func buildAuthURL(cfg *OAuthConfig, state, redirectURI string) string {
	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {redirectURI},
		"scope":         {cfg.Scope},
		"response_type": {"code"},
		"state":         {state},
	}
	return fmt.Sprintf("https://%s/oauth2/authorization?%s", cfg.Domain, params.Encode())
}

// exchangeCode は認可コードをアクセストークンに交換する
func exchangeCode(cfg *OAuthConfig, code, redirectURI string) (string, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", cfg.Domain)

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: %s", string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// トークンをキャッシュに保存
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Scope:        tokenResp.Scope,
		Expiry:       expiry.Format(time.RFC3339),
	}

	// --refresh-token が無効の場合はリフレッシュトークンを保存しない
	if !cfg.UseRefresh {
		token.RefreshToken = ""
	}

	if err := SaveToken(cfg.Domain, token); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache token: %v\n", err)
	}

	return tokenResp.AccessToken, nil
}

// refreshAccessToken はリフレッシュトークンを使ってアクセストークンを更新する
func refreshAccessToken(cfg *OAuthConfig, refreshToken string) (string, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", cfg.Domain)

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"refresh_token": {refreshToken},
	}

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh token request failed: %s", string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// トークンをキャッシュに保存
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Scope:        tokenResp.Scope,
		Expiry:       expiry.Format(time.RFC3339),
	}

	if err := SaveToken(cfg.Domain, token); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache token: %v\n", err)
	}

	return tokenResp.AccessToken, nil
}

// GetAuthURL は認可 URL を取得するためのヘルパー関数
func GetAuthURL(cfg *OAuthConfig) (string, string, error) {
	state, err := generateState()
	if err != nil {
		return "", "", err
	}
	redirectURI := fmt.Sprintf("https://localhost:%d/oauth2callback", cfg.Port)
	return buildAuthURL(cfg, state, redirectURI), state, nil
}

// openBrowser は OS に応じたブラウザを開く
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
