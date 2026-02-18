package kakao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/config"
	"github.com/qj0r9j0vc2/kko/internal/output"
)

const (
	authBaseURL   = "https://kauth.kakao.com"
	authorizePath = "/oauth/authorize"
	tokenPath     = "/oauth/token"
)

type Authenticator struct {
	cfg   *config.Config
	store config.CredentialStore
}

func NewAuthenticator(cfg *config.Config, store config.CredentialStore) *Authenticator {
	return &Authenticator{cfg: cfg, store: store}
}

func (a *Authenticator) EnsureValidToken(ctx context.Context) (string, error) {
	token, err := a.store.LoadToken()
	if err != nil {
		return "", fmt.Errorf("load token: %w", err)
	}
	if token == nil {
		return "", ErrNotAuthenticated
	}
	if token.ExpiresAt.After(time.Now().Add(5 * time.Minute)) {
		return token.AccessToken, nil
	}
	newToken, err := a.refreshToken(ctx, token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("token refresh failed: %w (run `kko auth login`)", err)
	}
	if err := a.store.SaveToken(newToken); err != nil {
		slog.Warn("failed to save refreshed token", "error", err)
	}
	return newToken.AccessToken, nil
}

func (a *Authenticator) Login(ctx context.Context) (*config.Token, error) {
	appKey := a.cfg.AppKey
	if appKey == "" {
		appKey = a.cfg.APIKey
	}
	if appKey == "" {
		return nil, ErrNoAPIKey
	}

	redirectURI := a.cfg.RedirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:9876/callback"
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect_uri: %w", err)
	}
	port := u.Port()
	if port == "" {
		port = "9876"
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			errCh <- fmt.Errorf("authorization failed: %s", errMsg)
			fmt.Fprintf(w, "<html><body><h2>Authorization failed</h2><p>%s</p><p>You can close this window.</p></body></html>", errMsg)
			return
		}
		codeCh <- code
		fmt.Fprint(w, "<html><body><h2>Success!</h2><p>You can close this window and return to the terminal.</p></body></html>")
	})

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server on port %s: %w", port, err)
	}

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	defer server.Shutdown(ctx)

	authURL := fmt.Sprintf("%s%s?client_id=%s&redirect_uri=%s&response_type=code&scope=talk_message,talk_calendar",
		authBaseURL, authorizePath,
		url.QueryEscape(appKey),
		url.QueryEscape(redirectURI),
	)

	fmt.Printf("Opening browser for Kakao login...\n")
	fmt.Printf("If browser doesn't open, visit:\n%s\n\n", authURL)
	output.OpenURL(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("login timed out after 5 minutes")
	}

	token, err := a.exchangeCode(ctx, code, appKey, redirectURI)
	if err != nil {
		return nil, err
	}

	if err := a.store.SaveToken(token); err != nil {
		slog.Warn("failed to save token", "error", err)
	}

	return token, nil
}

func (a *Authenticator) exchangeCode(_ context.Context, code, appKey, redirectURI string) (*config.Token, error) {
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {appKey},
		"redirect_uri": {redirectURI},
		"code":         {code},
	}
	if a.cfg.ClientSecret != "" {
		form.Set("client_secret", a.cfg.ClientSecret)
	}

	resp, err := http.PostForm(authBaseURL+tokenPath, form)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	scopes := splitScopes(tokenResp.Scope)

	return &config.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scopes:       scopes,
		TokenType:    tokenResp.TokenType,
	}, nil
}

func (a *Authenticator) refreshToken(_ context.Context, refreshToken string) (*config.Token, error) {
	appKey := a.cfg.AppKey
	if appKey == "" {
		appKey = a.cfg.APIKey
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {appKey},
		"refresh_token": {refreshToken},
	}
	if a.cfg.ClientSecret != "" {
		form.Set("client_secret", a.cfg.ClientSecret)
	}

	resp, err := http.PostForm(authBaseURL+tokenPath, form)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse refresh response: %w", err)
	}

	rt := tokenResp.RefreshToken
	if rt == "" {
		rt = refreshToken
	}

	scopes := splitScopes(tokenResp.Scope)

	return &config.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: rt,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scopes:       scopes,
		TokenType:    tokenResp.TokenType,
	}, nil
}

func (a *Authenticator) Status() (*config.Token, error) {
	return a.store.LoadToken()
}

func (a *Authenticator) Logout() error {
	return a.store.DeleteToken()
}

func splitScopes(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' || c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

