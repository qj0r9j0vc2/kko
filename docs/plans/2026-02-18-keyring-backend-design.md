# Keyring-Backend Credential Storage Design

**Date:** 2026-02-18
**Status:** Approved
**Author:** Engineering Team

## Problem Statement

Currently, sensitive credentials (`api_key`, `app_key`, `client_secret`) are stored in plain text in `~/.config/kko/config.yaml`. This poses a security risk:

- Credentials are readable by any process running under the same user
- Config files can be accidentally committed to version control
- No encryption or access control on credential storage
- Inconsistent with the existing OAuth token storage (which already uses keyring)

Meanwhile, OAuth tokens are already stored securely via system keyring with file fallback, demonstrating that the infrastructure exists but isn't being used for static credentials.

## Goals

1. **Security**: Store all sensitive credentials in system keyring (encrypted, access-controlled)
2. **Consistency**: Use same credential storage pattern for static keys and OAuth tokens
3. **Usability**: Provide fallback for environments where keyring is unavailable
4. **Explicit Security**: Make credential management commands clearly security-focused
5. **Breaking Change Acceptable**: Clean break preferred over complex migration

## Design Overview

### Architecture

```
┌─────────────────────────────────────────────────────┐
│                  CLI Commands                        │
│  (auth.go, config.go, calendar.go, etc.)           │
└─────────────────┬───────────────────────────────────┘
                  │
                  ├─ Non-sensitive config
                  │  (via Config struct)
                  ▼
      ┌─────────────────────┐
      │   Config (Viper)    │
      │  ~/.config/kko/     │
      │    config.yaml      │
      └─────────────────────┘
      • redirect_uri
      • output settings
      • aliases
      • search config

                  │
                  ├─ Sensitive credentials
                  │  (via CredentialStore)
                  ▼
      ┌─────────────────────────────────────┐
      │      CredentialStore Interface       │
      │  (extended with static creds)        │
      └────────────┬────────────────────────┘
                   │
           ┌───────┴────────┐
           │                │
           ▼                ▼
    ┌──────────┐    ┌─────────────┐
    │ Keyring  │    │ FileStore   │
    │  Store   │    │  (fallback) │
    └──────────┘    └─────────────┘
    System keyring   config.yaml
    - api-key        (read-only)
    - client-secret
    - oauth-token
```

### Key Principles

1. **Write-to-keyring-only**: All credential saves go to keyring; no programmatic file writes
2. **Read-with-fallback**: Try keyring first, fall back to config.yaml for reads
3. **Fetch-on-demand**: No credential caching in Config struct
4. **Explicit commands**: New `kko auth set-*` commands make security intent clear

## Detailed Design

### 1. Extended CredentialStore Interface

```go
// internal/config/store.go
type CredentialStore interface {
    // Existing token methods (unchanged)
    LoadToken() (*Token, error)
    SaveToken(token *Token) error
    DeleteToken() error

    // New static credential methods
    LoadAPIKey() (string, error)
    SaveAPIKey(key string) error
    LoadClientSecret() (string, error)
    SaveClientSecret(secret string) error
}
```

**KeyringStore Implementation:**
- Uses `github.com/zalando/go-keyring` (already a dependency)
- Service: `kko-kakao` (existing)
- Keys:
  - `api-key` → API key value
  - `client-secret` → Client secret value
  - `oauth-token` → Token JSON (existing)

**FileStore Implementation (Read-only for credentials):**
- `LoadAPIKey()` / `LoadClientSecret()`: Read from `config.yaml` if exists
- `SaveAPIKey()` / `SaveClientSecret()`: Return error with manual edit instructions
- Provides backward compatibility for reading existing configs
- Never writes credentials to files programmatically

### 2. Modified Config Struct

```go
// internal/config/config.go
type Config struct {
    // REMOVED: APIKey, AppKey, ClientSecret

    // Kept: non-sensitive settings
    RedirectURI     string            `mapstructure:"redirect_uri"`
    DefaultLocation Location          `mapstructure:"default_location"`
    Output          OutputConfig      `mapstructure:"output"`
    Aliases         map[string]string `mapstructure:"aliases"`
    Search          SearchConfig      `mapstructure:"search"`
}
```

### 3. Component Initialization

```go
// cmd/root.go
var (
    cfg    *config.Config
    store  config.CredentialStore  // NEW: global credential store
    client *kakao.Client
)

PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    cfg, err = config.Load(cfgFile)
    if err != nil {
        return err
    }

    store = config.NewCredentialStoreWithFallback()  // NEW
    client = kakao.NewClient(cfg, store)             // Pass store
    return nil
}
```

### 4. New Auth Commands

**Command Structure:**

```bash
kko auth set-api-key [value]       # Set API key (hidden input if no arg)
kko auth set-client-secret [value] # Set client secret (hidden input if no arg)
kko auth status                    # Show credential status (masked)
kko auth login                     # OAuth login (uses stored credentials)
kko auth logout                    # Clear OAuth token
```

**Secure Input Handling:**

```go
// Use golang.org/x/term for hidden input
import "golang.org/x/term"

func promptForSecret(prompt string) (string, error) {
    fmt.Print(prompt)
    secret, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Println() // New line after hidden input
    return string(secret), err
}
```

**Example Usage:**

```
$ kko auth set-api-key
Enter API key: ••••••••••••  (hidden)
✓ API key saved to keyring

$ kko auth set-api-key "abc123"
✓ API key saved to keyring

$ kko auth status
API Key: abc***  (from keyring)
Client Secret: xyz***  (from keyring)
OAuth Token: Valid (expires in 2h)
```

### 5. Data Flow

**Setting Credentials:**

```
User → kko auth set-api-key
     → store.SaveAPIKey()
     → KeyringStore.SaveAPIKey()
     → keyring.Set("kko-kakao", "api-key", value)
     → Success or error with manual fallback instructions
```

**Using Credentials:**

```
API Request → client.do()
           → apiKey, err := store.LoadAPIKey()
           → Try keyring first
           → Fallback to config.yaml if needed
           → Add to Authorization header
           → Send request
```

### 6. Error Handling

**Error Types:**

```go
// internal/config/errors.go
var (
    ErrNoAPIKey           = errors.New("API key not configured: run `kko auth set-api-key`")
    ErrNoClientSecret     = errors.New("client secret not configured: run `kko auth set-client-secret`")
    ErrKeyringUnavailable = errors.New("keyring unavailable")
    ErrCredentialNotFound = errors.New("credential not found")
)
```

**Error Scenarios:**

| Scenario | Error Message | Recovery |
|----------|---------------|----------|
| No API key | `API key not configured` | `kko auth set-api-key` |
| No client secret (optional) | Silent - OAuth works without it | N/A |
| Keyring save failed | `Failed to save: <reason>. Fallback: edit config.yaml` | Manual config edit |
| Both keyring and config unavailable | `Cannot access credentials` | Setup instructions |

### 7. Backward Compatibility

**Breaking Changes:**
- Commands using `kko config set api_key` will no longer work
- Credentials in `config.yaml` will be read-only (won't be updated by CLI)
- Users must migrate to `kko auth set-api-key` for new credential setting

**Migration Path:**
1. Users with existing `config.yaml` credentials can continue using them (read-only fallback)
2. To move to secure storage: run `kko auth set-api-key` and `kko auth set-client-secret`
3. Old credentials in config.yaml can be manually removed by users after migration

## Implementation Plan

### Files to Modify

1. **`internal/config/store.go`**
   - Extend `CredentialStore` interface
   - Update `KeyringStore` with new methods
   - Update `FileStore` with read-only fallback
   - Add `NewCredentialStoreWithFallback()` factory

2. **`internal/config/config.go`**
   - Remove `APIKey`, `AppKey`, `ClientSecret` from `Config` struct
   - Keep `Load()` for non-sensitive config

3. **`internal/config/errors.go`** (new)
   - Define credential-related errors

4. **`cmd/auth.go`**
   - Add `set-api-key` command
   - Add `set-client-secret` command
   - Update `status` command to show credential sources
   - Modify `login` to use store

5. **`cmd/root.go`**
   - Initialize `CredentialStore` in persistent pre-run
   - Pass store to client initialization

6. **`internal/kakao/client.go`**
   - Accept `CredentialStore` in constructor
   - Fetch credentials on-demand from store

7. **`internal/kakao/auth.go`**
   - Load client secret from store when needed

### Testing Strategy

**Unit Tests:**
- `store_test.go`: Test all KeyringStore and FileStore methods
- `config_test.go`: Verify Config struct changes
- `client_test.go`: Mock CredentialStore usage

**Integration Tests:**
- End-to-end credential setting and usage
- Fallback behavior verification
- Error handling validation

**Manual Testing:**
```bash
# Set credentials
kko auth set-api-key
kko auth set-client-secret

# Verify
kko auth status

# Use
kko local search "test"
kko auth login

# Fallback test
# Manually create config.yaml with credentials
# Verify reading works when keyring unavailable
```

## Security Considerations

1. **Keyring Security**: System keyring provides OS-level encryption and access control
2. **Hidden Input**: Credentials never echoed to terminal during input
3. **No Logging**: Credentials not logged or exposed in error messages
4. **Masked Display**: `kko auth status` shows masked credentials (e.g., `abc***`)
5. **Read-only Fallback**: Never writes credentials to files programmatically

## Non-Goals

- Automatic migration from config.yaml to keyring (users do this manually)
- Encrypted file storage (keyring-only writes, plain YAML read fallback)
- Credential rotation or expiry management
- Multi-profile credential management

## Success Metrics

- [ ] All credentials stored in system keyring by default
- [ ] Clear error messages when credentials unavailable
- [ ] Fallback works for constrained environments
- [ ] No programmatic writes to config.yaml for credentials
- [ ] Hidden input for interactive credential entry
- [ ] All existing API and OAuth flows work with new storage

## Future Enhancements

- Credential rotation support
- Multiple profile management
- Credential expiry warnings
- Integration with cloud secret managers (AWS Secrets Manager, etc.)
