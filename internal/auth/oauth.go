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
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

const (
	configDir       = ".config/gt"
	credentialsFile = "credentials.json"
	tokenFile       = "token.json"
)

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

func getCredentialsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}

func getTokenPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFile), nil
}

func loadCredentials() (*oauth2.Config, error) {
	path, err := getCredentialsPath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("credentials.json not found: %s\nPlease download it from Google Cloud Console", path)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials.json: %w", err)
	}

	return config, nil
}

func loadToken() (*oauth2.Token, error) {
	path, err := getTokenPath()
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

func saveToken(token *oauth2.Token) error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := getTokenPath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// GetClient returns an authenticated HTTP client.
// If no token exists or it's expired, it automatically triggers browser authentication.
func GetClient(ctx context.Context) (*http.Client, error) {
	config, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	token, err := loadToken()
	if err != nil {
		// No token, authenticate via browser
		token, err = authenticateViaBrowser(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(token); err != nil {
			return nil, err
		}
	} else if token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
		// Token expired, try to refresh
		tokenSource := config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			// Refresh failed, re-authenticate via browser
			newToken, err = authenticateViaBrowser(ctx, config)
			if err != nil {
				return nil, err
			}
		}
		if err := saveToken(newToken); err != nil {
			return nil, err
		}
		token = newToken
	}

	return config.Client(ctx, token), nil
}

// authenticateViaBrowser performs OAuth authentication via browser
func authenticateViaBrowser(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Set redirect URL
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Channel to receive auth code
	codeCh := make(chan string)
	errCh := make(chan error)

	// Start local server for callback
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("failed to get authorization code")
			fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1><p>You can close this window.</p></body></html>")
			return
		}
		codeCh <- code
		fmt.Fprintf(w, "<html><body><h1>Authentication Successful!</h1><p>You can close this window.</p></body></html>")
	})
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Generate auth URL and open browser
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("Please authenticate in your browser...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser. Please open the following URL manually:\n%s\n", authURL)
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
		return nil, fmt.Errorf("failed to get token: %w", err)
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
