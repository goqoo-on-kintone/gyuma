package cert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/goqoo-on-kintone/gyuma/internal/config"
)

// MkcertStatus は mkcert のインストール状態を表す
type MkcertStatus struct {
	Installed bool
	Path      string
	CAInstalled bool
}

// CheckMkcert は mkcert のインストール状態を確認する
func CheckMkcert() (*MkcertStatus, error) {
	path, err := exec.LookPath("mkcert")
	if err != nil {
		return &MkcertStatus{Installed: false}, nil
	}

	// ルート CA がインストールされているか確認
	// mkcert -CAROOT でルート CA のパスを取得
	cmd := exec.Command(path, "-CAROOT")
	output, err := cmd.Output()
	if err != nil {
		return &MkcertStatus{Installed: true, Path: path, CAInstalled: false}, nil
	}

	caRoot := strings.TrimSpace(string(output))
	rootCertPath := filepath.Join(caRoot, "rootCA.pem")
	caInstalled := fileExists(rootCertPath)

	return &MkcertStatus{
		Installed:   true,
		Path:        path,
		CAInstalled: caInstalled,
	}, nil
}

// SetupMkcertCert は mkcert を使って証明書をセットアップする
func SetupMkcertCert(host string, force bool) error {
	status, err := CheckMkcert()
	if err != nil {
		return err
	}

	if !status.Installed {
		return fmt.Errorf("mkcert is not installed.\n\nInstall mkcert:\n%s", getInstallInstructions())
	}

	fmt.Fprintf(os.Stderr, "✓ mkcert found: %s\n", status.Path)

	// ルート CA をインストール
	if !status.CAInstalled {
		fmt.Fprintln(os.Stderr, "Installing root CA...")
		cmd := exec.Command(status.Path, "-install")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install root CA: %w", err)
		}
		fmt.Fprintln(os.Stderr, "✓ Root CA installed")
	} else {
		fmt.Fprintln(os.Stderr, "✓ Root CA already installed")
	}

	// 証明書ディレクトリを確認
	if err := config.EnsureCertsDir(); err != nil {
		return err
	}

	certPath, err := config.MkcertCertFile()
	if err != nil {
		return err
	}
	keyPath, err := config.MkcertKeyFile()
	if err != nil {
		return err
	}

	// 既存の証明書を確認
	if !force && fileExists(certPath) && fileExists(keyPath) {
		fmt.Fprintln(os.Stderr, "✓ Certificate already exists (use --force to regenerate)")
		return nil
	}

	// 証明書を生成
	certsDir, err := config.CertsDir()
	if err != nil {
		return err
	}

	// mkcert は出力ファイル名を自動で決めるので、一時ディレクトリで生成して移動する
	cmd := exec.Command(status.Path,
		"-cert-file", certPath,
		"-key-file", keyPath,
		host,
	)
	cmd.Dir = certsDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Certificate generated: %s\n", certPath)
	fmt.Fprintln(os.Stderr, "Setup complete! You can now use gyuma without browser warnings.")

	return nil
}

// getInstallInstructions は OS に応じたインストール手順を返す
func getInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return "  brew install mkcert\n  brew install nss  # if you use Firefox"
	case "linux":
		return "  # Debian/Ubuntu:\n  sudo apt install mkcert\n\n  # Arch Linux:\n  sudo pacman -S mkcert\n\n  # Other:\n  go install filippo.io/mkcert@latest"
	case "windows":
		return "  choco install mkcert\n  # or\n  scoop install mkcert"
	default:
		return "  go install filippo.io/mkcert@latest"
	}
}

// HasMkcertCert は mkcert 証明書が存在するかどうかを判定する
func HasMkcertCert() bool {
	certPath, err := config.MkcertCertFile()
	if err != nil {
		return false
	}
	keyPath, err := config.MkcertKeyFile()
	if err != nil {
		return false
	}
	return fileExists(certPath) && fileExists(keyPath)
}
