package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/goqoo-on-kintone/gyuma/internal/config"
)

const (
	// SelfSignedValidDays は自己署名証明書の有効期間（日数）
	SelfSignedValidDays = 30
)

// CertPaths は証明書と秘密鍵のパスを保持する
type CertPaths struct {
	CertFile string
	KeyFile  string
}

// GetCertPaths は利用可能な証明書のパスを返す
// mkcert 証明書を優先し、なければ自己署名証明書を返す
func GetCertPaths() (*CertPaths, bool, error) {
	// mkcert 証明書を確認
	mkcertCert, err := config.MkcertCertFile()
	if err != nil {
		return nil, false, err
	}
	mkcertKey, err := config.MkcertKeyFile()
	if err != nil {
		return nil, false, err
	}

	if fileExists(mkcertCert) && fileExists(mkcertKey) {
		return &CertPaths{
			CertFile: mkcertCert,
			KeyFile:  mkcertKey,
		}, true, nil
	}

	// 自己署名証明書を確認・生成
	selfCert, err := config.SelfSignedCertFile()
	if err != nil {
		return nil, false, err
	}
	selfKey, err := config.SelfSignedKeyFile()
	if err != nil {
		return nil, false, err
	}

	// 自己署名証明書の有効期限を確認
	needsRegen := false
	if !fileExists(selfCert) || !fileExists(selfKey) {
		needsRegen = true
	} else {
		expired, err := isCertExpiringSoon(selfCert)
		if err != nil || expired {
			needsRegen = true
		}
	}

	if needsRegen {
		if err := generateSelfSignedCert(); err != nil {
			return nil, false, fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
	}

	return &CertPaths{
		CertFile: selfCert,
		KeyFile:  selfKey,
	}, false, nil
}

// generateSelfSignedCert は自己署名証明書を生成する
func generateSelfSignedCert() error {
	if err := config.EnsureCertsDir(); err != nil {
		return err
	}

	// ECDSA P-256 秘密鍵を生成
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// 証明書テンプレートを作成
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gyuma OAuth"},
			CommonName:   "localhost",
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, SelfSignedValidDays),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// 自己署名証明書を作成
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// 証明書をファイルに保存
	certPath, err := config.SelfSignedCertFile()
	if err != nil {
		return err
	}
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write cert file: %w", err)
	}

	// 秘密鍵をファイルに保存
	keyPath, err := config.SelfSignedKeyFile()
	if err != nil {
		return err
	}
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	// パーミッションを設定
	if err := os.Chmod(certPath, 0600); err != nil {
		return fmt.Errorf("failed to set cert file permissions: %w", err)
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %w", err)
	}

	return nil
}

// isCertExpiringSoon は証明書が期限切れまたは期限間近かどうかを判定する
func isCertExpiringSoon(certPath string) (bool, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return true, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return true, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true, err
	}

	// 有効期限まで1日未満なら再生成
	return time.Until(cert.NotAfter) < 24*time.Hour, nil
}

// fileExists はファイルが存在するかどうかを判定する
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
