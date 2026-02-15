package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

const (
	configDir       = ".config/gt"
	credentialsFile = "credentials.json"
	tokenFile       = "token.json"
)

// GetConfigDir returns the path to the config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

// GetCredentialsPath returns the path to the credentials file
func GetCredentialsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}

// GetTokenPath returns the path to the token file
func GetTokenPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFile), nil
}

// LoadCredentials loads OAuth credentials from the config file
func LoadCredentials() (*oauth2.Config, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("credentials.json が見つかりません: %s\nGoogle Cloud Console からダウンロードして配置してください", path)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return nil, fmt.Errorf("credentials.json のパースに失敗しました: %w", err)
	}

	return config, nil
}

// LoadToken loads a saved token from the token file
func LoadToken() (*oauth2.Token, error) {
	path, err := GetTokenPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// SaveToken saves a token to the token file
func SaveToken(token *oauth2.Token) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	path, err := GetTokenPath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("トークンファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// DeleteToken deletes the saved token file
func DeleteToken() error {
	path, err := GetTokenPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Already logged out
		}
		return fmt.Errorf("トークンの削除に失敗しました: %w", err)
	}

	return nil
}

// HasToken checks if a valid token exists
func HasToken() bool {
	_, err := LoadToken()
	return err == nil
}

// GetClient returns an authenticated HTTP client
func GetClient(ctx context.Context) (*http.Client, error) {
	config, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	token, err := LoadToken()
	if err != nil {
		return nil, fmt.Errorf("ログインが必要です。`gt login` を実行してください")
	}

	return config.Client(ctx, token), nil
}

// Authenticate performs OAuth authentication via browser
func Authenticate(ctx context.Context) (*oauth2.Token, error) {
	config, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	// Find an available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("ローカルサーバーの起動に失敗しました: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Set redirect URL
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Channel to receive auth code
	codeCh := make(chan string)
	errCh := make(chan error)

	// Start local server for callback
	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port)}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("認証コードを取得できませんでした")
			fmt.Fprintf(w, "<html><body><h1>認証失敗</h1><p>ウィンドウを閉じてください。</p></body></html>")
			return
		}
		codeCh <- code
		fmt.Fprintf(w, "<html><body><h1>認証成功!</h1><p>このウィンドウを閉じてください。</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Generate auth URL and open browser
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("ブラウザで認証してください...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("ブラウザを開けませんでした。以下のURLを手動で開いてください:\n%s\n", authURL)
	}

	// Wait for auth code
	var authCode string
	select {
	case authCode = <-codeCh:
	case err := <-errCh:
		server.Shutdown(ctx)
		return nil, err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return nil, ctx.Err()
	}

	server.Shutdown(ctx)

	// Exchange auth code for token
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("トークンの取得に失敗しました: %w", err)
	}

	return token, nil
}

// openBrowser opens the specified URL in the default browser
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
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
