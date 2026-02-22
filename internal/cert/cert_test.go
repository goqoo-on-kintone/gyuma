package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileExists(t *testing.T) {
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "gyuma-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "存在するファイル",
			path: tmpPath,
			want: true,
		},
		{
			name: "存在しないファイル",
			path: "/nonexistent/path/to/file",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileExists(tt.path); got != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateSelfSignedCert(t *testing.T) {
	// テスト用の一時ディレクトリを使用するため、スキップ
	// 実際の ~/.config/gyuma/certs に書き込むため、統合テストで確認
	t.Skip("This test writes to real config directory - run manually if needed")
}

func TestIsCertExpiringSoon(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "gyuma-cert-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// ECDSA キーを生成
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// 有効な証明書を作成（テスト用に簡易的に）
	validCertPath := filepath.Join(tmpDir, "valid.pem")
	expiredCertPath := filepath.Join(tmpDir, "expired.pem")

	// テスト用の証明書データを作成
	createTestCert := func(path string, notAfter time.Time) error {
		// 最小限の自己署名証明書を生成
		template := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     notAfter,
		}

		certDER, err := x509.CreateCertificate(
			rand.Reader, template, template, &ecdsaKey.PublicKey, ecdsaKey,
		)
		if err != nil {
			return err
		}

		certFile, err := os.Create(path)
		if err != nil {
			return err
		}
		defer certFile.Close()

		return pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	}

	// 有効な証明書（期限まで1週間）
	if err := createTestCert(validCertPath, time.Now().Add(7*24*time.Hour)); err != nil {
		t.Fatalf("Failed to create valid cert: %v", err)
	}

	// 期限切れ間近の証明書（期限まで1時間）
	if err := createTestCert(expiredCertPath, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("Failed to create expired cert: %v", err)
	}

	tests := []struct {
		name     string
		certPath string
		want     bool
		wantErr  bool
	}{
		{
			name:     "有効な証明書",
			certPath: validCertPath,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "期限切れ間近の証明書",
			certPath: expiredCertPath,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "存在しないファイル",
			certPath: "/nonexistent/cert.pem",
			want:     true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isCertExpiringSoon(tt.certPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("isCertExpiringSoon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isCertExpiringSoon() = %v, want %v", got, tt.want)
			}
		})
	}
}
