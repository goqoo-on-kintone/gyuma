package config

import (
	"fmt"
	"io"

	"github.com/bgentry/go-netrc/netrc"
)

// netrcMachineName は domain に :oauth suffix を付けた machine 名を返す。
// kintone ユーザー認証（suffix なし）や Basic 認証（:basic）と用途を分離する。
func netrcMachineName(domain string) string {
	return domain + ":oauth"
}

// findOAuthCredentials は netrc の内容から <domain>:oauth machine の login/password を
// client_id/client_secret として取り出す。該当 machine が無ければエラーを返す。
func findOAuthCredentials(r io.Reader, domain string) (*Credentials, error) {
	n, err := netrc.Parse(r)
	if err != nil {
		return nil, err
	}

	name := netrcMachineName(domain)
	m := n.FindMachine(name)
	// FindMachine は一致が無いと default machine を返すため除外する
	if m == nil || m.IsDefault() || m.Login == "" || m.Password == "" {
		return nil, fmt.Errorf("no oauth credentials for %q in netrc", name)
	}

	return &Credentials{
		ClientID:     m.Login,
		ClientSecret: m.Password,
	}, nil
}
