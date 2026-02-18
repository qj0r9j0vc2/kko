package kakao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/config"
	"golang.org/x/time/rate"
)

type Client struct {
	http        *http.Client
	store       config.CredentialStore
	auth        *Authenticator
	rateLimiter *rate.Limiter
	logger      *slog.Logger
	cfg         *config.Config
}

func NewClient(cfg *config.Config, store config.CredentialStore) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		store:       store,
		auth:        NewAuthenticator(cfg, store),
		rateLimiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 10),
		logger:      slog.Default(),
		cfg:         cfg,
	}
}

type requestConfig struct {
	useOAuth bool
	method   string
	body     io.Reader
}

func defaultRequestConfig() *requestConfig {
	return &requestConfig{method: "GET"}
}

type RequestOption func(*requestConfig)

func WithOAuth() RequestOption {
	return func(c *requestConfig) {
		c.useOAuth = true
	}
}

func (c *Client) Get(ctx context.Context, endpoint string, params url.Values, opts ...RequestOption) ([]byte, error) {
	cfg := defaultRequestConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.method = "GET"

	fullURL := endpoint
	if len(params) > 0 {
		fullURL = endpoint + "?" + params.Encode()
	}

	return c.doWithRetry(ctx, cfg, fullURL, nil)
}

func (c *Client) Post(ctx context.Context, endpoint string, form url.Values, opts ...RequestOption) ([]byte, error) {
	cfg := defaultRequestConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.method = "POST"
	cfg.body = strings.NewReader(form.Encode())

	return c.doWithRetry(ctx, cfg, endpoint, http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	})
}

func (c *Client) PostJSON(ctx context.Context, endpoint string, payload interface{}, opts ...RequestOption) ([]byte, error) {
	cfg := defaultRequestConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.method = "POST"

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	cfg.body = strings.NewReader(string(data))

	return c.doWithRetry(ctx, cfg, endpoint, http.Header{
		"Content-Type": {"application/json"},
	})
}

func (c *Client) Put(ctx context.Context, endpoint string, form url.Values, opts ...RequestOption) ([]byte, error) {
	cfg := defaultRequestConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.method = "PUT"
	cfg.body = strings.NewReader(form.Encode())

	return c.doWithRetry(ctx, cfg, endpoint, http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	})
}

func (c *Client) Delete(ctx context.Context, endpoint string, params url.Values, opts ...RequestOption) ([]byte, error) {
	cfg := defaultRequestConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.method = "DELETE"

	fullURL := endpoint
	if len(params) > 0 {
		fullURL = endpoint + "?" + params.Encode()
	}

	return c.doWithRetry(ctx, cfg, fullURL, nil)
}

func (c *Client) doWithRetry(ctx context.Context, cfg *requestConfig, fullURL string, extraHeaders http.Header) ([]byte, error) {
	maxRetries := 3
	baseDelay := 500 * time.Millisecond
	maxDelay := 10 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}
			c.logger.Debug("retrying request", "attempt", attempt, "delay", delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		body, err := c.do(ctx, cfg, fullURL, extraHeaders)
		if err == nil {
			return body, nil
		}

		lastErr = err
		apiErr, ok := err.(*APIError)
		if !ok {
			return nil, err
		}

		if apiErr.Status == http.StatusTooManyRequests || apiErr.Status >= 500 {
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) do(ctx context.Context, cfg *requestConfig, fullURL string, extraHeaders http.Header) ([]byte, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, cfg.method, fullURL, cfg.body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if cfg.useOAuth {
		token, err := c.auth.EnsureValidToken(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		// Fetch API key from store on-demand
		apiKey, err := c.store.LoadAPIKey()
		if err != nil {
			return nil, config.ErrNoAPIKey
		}
		req.Header.Set("Authorization", "KakaoAK "+apiKey)
	}

	for k, v := range extraHeaders {
		req.Header[k] = v
	}

	c.logger.Debug("request", "method", cfg.method, "url", fullURL)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		apiErr.Status = resp.StatusCode
		return nil, &apiErr
	}

	return body, nil
}

func (c *Client) Config() *config.Config {
	return c.cfg
}

func (c *Client) Auth() *Authenticator {
	return c.auth
}
