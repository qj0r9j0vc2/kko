package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

type FileStore struct {
	path string
}

func NewFileStore() *FileStore {
	return &FileStore{
		path: filepath.Join(DefaultConfigDir(), "token.json"),
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
