package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goqoo-on-kintone/gyuma/internal/auth"
	"github.com/goqoo-on-kintone/gyuma/internal/cert"
	"github.com/goqoo-on-kintone/gyuma/internal/config"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup-cert" {
		if err := runSetupCert(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMain() error {
	// フラグ定義
	var (
		domain       string
		clientID     string
		clientSecret string
		scope        string
		port         int
		useRefresh   bool
		proxy        string
		pfxFilepath  string
		pfxPassword  string
		noPrompt     bool
		quiet        bool
		showVersion  bool
		showHelp     bool
	)

	flag.StringVar(&domain, "d", "", "kintone domain name")
	flag.StringVar(&domain, "domain", "", "kintone domain name")
	flag.StringVar(&clientID, "i", "", "OAuth2 Client ID")
	flag.StringVar(&clientID, "client-id", "", "OAuth2 Client ID")
	flag.StringVar(&clientSecret, "s", "", "OAuth2 Client Secret")
	flag.StringVar(&clientSecret, "client-secret", "", "OAuth2 Client Secret")
	flag.StringVar(&scope, "S", "", "OAuth2 Scope (space or comma separated)")
	flag.StringVar(&scope, "scope", "", "OAuth2 Scope (space or comma separated)")
	flag.IntVar(&port, "P", 3000, "Local server port")
	flag.IntVar(&port, "port", 3000, "Local server port")
	flag.BoolVar(&useRefresh, "refresh-token", false, "Save and use refresh token")
	flag.StringVar(&proxy, "proxy", "", "Proxy server")
	flag.StringVar(&pfxFilepath, "pfx-filepath", "", "Client certificate file path")
	flag.StringVar(&pfxPassword, "pfx-password", "", "Client certificate password")
	flag.BoolVar(&noPrompt, "noprompt", false, "Disable interactive input")
	flag.BoolVar(&quiet, "quiet", false, "Suppress warning messages")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Parse()

	if showHelp {
		printUsage()
		return nil
	}

	if showVersion {
		fmt.Println(version)
		return nil
	}

	// 必須パラメータの検証
	if domain == "" {
		return fmt.Errorf("--domain is required")
	}
	if scope == "" {
		return fmt.Errorf("--scope is required")
	}

	// クレデンシャルを取得
	creds, source, err := config.GetCredentials(domain, clientID, clientSecret, noPrompt)
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Using credentials from %s\n", source)
	}

	// OAuth フローを実行
	oauthConfig := &auth.OAuthConfig{
		Domain:       domain,
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Scope:        scope,
		Port:         port,
		UseRefresh:   useRefresh,
		Quiet:        quiet,
	}

	// キャッシュされたトークンを確認
	token, err := auth.GetValidToken(domain, scope)
	if err != nil {
		return err
	}

	if token != nil {
		// 有効なトークンがある場合はそのまま出力
		fmt.Println(token.AccessToken)
		return nil
	}

	// 新規取得が必要な場合は OAuth フローを実行
	accessToken, err := auth.RunOAuthFlow(oauthConfig)
	if err != nil {
		return err
	}

	// アクセストークンを stdout に出力
	fmt.Println(accessToken)
	return nil
}

func runSetupCert(args []string) error {
	fs := flag.NewFlagSet("setup-cert", flag.ExitOnError)
	var (
		host  string
		port  int
		force bool
	)
	fs.StringVar(&host, "host", "localhost", "Certificate hostname")
	fs.IntVar(&port, "port", 3000, "HTTPS server port")
	fs.BoolVar(&force, "force", false, "Overwrite existing certificate")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return cert.SetupMkcertCert(host, force)
}

func printUsage() {
	fmt.Println(`gyuma - kintone OAuth 2.0 CLI tool

Usage:
  gyuma [options]                    Get OAuth access token
  gyuma setup-cert [options]         Setup mkcert certificate

Main command options:
  -d, --domain           kintone domain name (required)
  -i, --client-id        OAuth2 Client ID
  -s, --client-secret    OAuth2 Client Secret
  -S, --scope            OAuth2 Scope (required, space or comma separated)
  -P, --port             Local server port (default: 3000)
      --refresh-token    Save and use refresh token (default: disabled)
      --proxy            Proxy server
      --pfx-filepath     Client certificate file path
      --pfx-password     Client certificate password
      --noprompt         Disable interactive input
      --quiet            Suppress warning messages
  -v, --version          Show version
  -h, --help             Show help

setup-cert options:
      --host             Certificate hostname (default: localhost)
      --port             HTTPS server port (default: 3000)
      --force            Overwrite existing certificate

Environment variables:
  GYUMA_CLIENT_ID        OAuth2 Client ID
  GYUMA_CLIENT_SECRET    OAuth2 Client Secret

Examples:
  # Get access token
  gyuma -d example.cybozu.com -S "k:app_settings:read"

  # Setup mkcert certificate
  gyuma setup-cert`)
}
