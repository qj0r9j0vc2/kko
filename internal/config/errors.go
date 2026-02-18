package config

import "errors"

var (
	ErrNoAPIKey           = errors.New("API key not configured: run `kko auth set-api-key`")
	ErrNoClientSecret     = errors.New("client secret not configured: run `kko auth set-client-secret`")
	ErrKeyringUnavailable = errors.New("keyring unavailable")
	ErrCredentialNotFound = errors.New("credential not found")
)
