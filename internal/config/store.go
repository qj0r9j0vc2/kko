package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "kko-kakao"
	keyringUser    = "oauth-token"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes"`
	TokenType    string    `json:"token_type"`
}

type CredentialStore interface {
	LoadToken() (*Token, error)
	SaveToken(token *Token) error
	DeleteToken() error
	LoadAPIKey() (string, error)
	SaveAPIKey(key string) error
	LoadClientSecret() (string, error)
	SaveClientSecret(secret string) error
}

type KeyringStore struct{}

func NewKeyringStore() *KeyringStore {
	return &KeyringStore{}
}

func (s *KeyringStore) LoadToken() (*Token, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}
		return NewFileStore().LoadToken()
	}
	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("corrupt token in keyring: %w", err)
	}
	return &token, nil
}

func (s *KeyringStore) SaveToken(token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return NewFileStore().SaveToken(token)
	}
	return nil
}

func (s *KeyringStore) DeleteToken() error {
	err := keyring.Delete(keyringService, keyringUser)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	_ = NewFileStore().DeleteToken()
	return nil
}

func (s *KeyringStore) LoadAPIKey() (string, error) {
	key, err := keyring.Get(keyringService, "api-key")
	if err != nil {
		if err == keyring.ErrNotFound {
			return NewFileStore().LoadAPIKey()
		}
		return "", err
	}
	return key, nil
}

func (s *KeyringStore) SaveAPIKey(key string) error {
	if err := keyring.Set(keyringService, "api-key", key); err != nil {
		return fmt.Errorf("failed to save API key to keyring: %w", err)
	}
	return nil
}

func (s *KeyringStore) LoadClientSecret() (string, error) {
	secret, err := keyring.Get(keyringService, "client-secret")
	if err != nil {
		if err == keyring.ErrNotFound {
			return NewFileStore().LoadClientSecret()
		}
		return "", err
	}
	return secret, nil
}

func (s *KeyringStore) SaveClientSecret(secret string) error {
	if err := keyring.Set(keyringService, "client-secret", secret); err != nil {
		return fmt.Errorf("failed to save client secret to keyring: %w", err)
	}
	return nil
}

type FileStore struct {
	path       string
	configPath string
}

func NewFileStore() *FileStore {
	return &FileStore{
		path:       filepath.Join(DefaultConfigDir(), "token.json"),
		configPath: DefaultConfigPath(),
	}
}

func (s *FileStore) LoadToken() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("corrupt token file: %w", err)
	}
	return &token, nil
}

func (s *FileStore) SaveToken(token *Token) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *FileStore) DeleteToken() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FileStore) loadConfigValue(key string) (string, error) {
	v := viper.New()
	v.SetConfigFile(s.configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	value := v.GetString(key)
	if value == "" {
		return "", fmt.Errorf("%s not found in config file", key)
	}

	return value, nil
}

func (s *FileStore) LoadAPIKey() (string, error) {
	// Try api_key first, then app_key as fallback
	if key, err := s.loadConfigValue("api_key"); err == nil {
		return key, nil
	}
	return s.loadConfigValue("app_key")
}

func (s *FileStore) SaveAPIKey(key string) error {
	return fmt.Errorf("keyring unavailable: please manually edit %s to add 'api_key: <your-key>'", s.configPath)
}

func (s *FileStore) LoadClientSecret() (string, error) {
	return s.loadConfigValue("client_secret")
}

func (s *FileStore) SaveClientSecret(secret string) error {
	return fmt.Errorf("keyring unavailable: please manually edit %s to add 'client_secret: <your-secret>'", s.configPath)
}
