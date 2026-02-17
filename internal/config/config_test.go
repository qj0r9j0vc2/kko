package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := &FileStore{path: filepath.Join(dir, "token.json")}

	token := &Token{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now().Add(time.Hour).Truncate(time.Second),
		Scopes:       []string{"talk_message"},
		TokenType:    "bearer",
	}

	if err := store.SaveToken(token); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadToken()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, token.RefreshToken)
	}
	if len(loaded.Scopes) != 1 || loaded.Scopes[0] != "talk_message" {
		t.Errorf("Scopes = %v, want [talk_message]", loaded.Scopes)
	}
}

func TestFileStore_LoadToken_NotExist(t *testing.T) {
	store := &FileStore{path: filepath.Join(t.TempDir(), "nonexistent.json")}
	token, err := store.LoadToken()
	if err != nil {
		t.Fatal(err)
	}
	if token != nil {
		t.Error("expected nil token for nonexistent file")
	}
}

func TestFileStore_DeleteToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	store := &FileStore{path: path}
	if err := store.DeleteToken(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestFileStore_DeleteToken_NotExist(t *testing.T) {
	store := &FileStore{path: filepath.Join(t.TempDir(), "nonexistent.json")}
	err := store.DeleteToken()
	if err != nil {
		t.Errorf("deleting nonexistent token should not error, got %v", err)
	}
}

func TestFileStore_SaveToken_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	store := &FileStore{path: filepath.Join(dir, "token.json")}

	token := &Token{AccessToken: "test"}
	if err := store.SaveToken(token); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "token.json")); os.IsNotExist(err) {
		t.Error("expected token file to be created in nested dir")
	}
}
