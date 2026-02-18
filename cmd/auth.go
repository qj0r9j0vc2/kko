package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Kakao OAuth authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with Kakao account",
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE:  runAuthLogout,
}

var authSetAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key [key]",
	Short: "Set Kakao API key (stored in keyring)",
	RunE:  runAuthSetAPIKey,
}

var authSetClientSecretCmd = &cobra.Command{
	Use:   "set-client-secret [secret]",
	Short: "Set Kakao client secret (stored in keyring)",
	RunE:  runAuthSetClientSecret,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authSetAPIKeyCmd)
	authCmd.AddCommand(authSetClientSecretCmd)
}

func runAuthLogin(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	token, err := client.Auth().Login(ctx)
	if err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]interface{}{
			"status":     "authenticated",
			"expires_at": token.ExpiresAt,
			"scopes":     token.Scopes,
		})
	}

	fmt.Println()
	fmt.Printf("  %s Logged in successfully!\n", output.Success("✓"))
	fmt.Printf("  %s %s\n", output.Label("Scopes:"), strings.Join(token.Scopes, ", "))
	fmt.Printf("  %s %s\n\n", output.Label("Expires:"), token.ExpiresAt.Format("2006-01-02 15:04"))
	return nil
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	token, err := client.Auth().Status()
	if err != nil {
		return err
	}

	// Fetch API key from store
	apiKeyStatus := "not configured"
	apiKey, err := store.LoadAPIKey()
	if err == nil && apiKey != "" {
		masked := apiKey
		if len(masked) > 4 {
			masked = "..." + masked[len(masked)-4:]
		}
		apiKeyStatus = fmt.Sprintf("configured (ends with %s)", masked)
	}

	// Fetch client secret from store
	clientSecretStatus := "not configured"
	clientSecret, err := store.LoadClientSecret()
	if err == nil && clientSecret != "" {
		masked := clientSecret
		if len(masked) > 4 {
			masked = "..." + masked[len(masked)-4:]
		}
		clientSecretStatus = fmt.Sprintf("configured (ends with %s)", masked)
	}

	oauthStatus := "not authenticated"
	tokenStatus := "none"
	scopeStr := ""
	if token != nil {
		oauthStatus = "authenticated"
		if token.ExpiresAt.After(time.Now()) {
			remaining := time.Until(token.ExpiresAt)
			hours := int(remaining.Hours())
			mins := int(remaining.Minutes()) % 60
			tokenStatus = fmt.Sprintf("valid (expires in %dh %dm)", hours, mins)
		} else {
			tokenStatus = "expired"
		}
		scopeStr = strings.Join(token.Scopes, ", ")
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]interface{}{
			"api_key":       apiKeyStatus,
			"client_secret": clientSecretStatus,
			"oauth":         oauthStatus,
			"token":         tokenStatus,
			"scopes":        scopeStr,
		})
	}

	fmt.Println()
	fmt.Printf("  %s         %s\n", output.Label("API Key:"), apiKeyStatus)
	fmt.Printf("  %s  %s\n", output.Label("Client Secret:"), clientSecretStatus)
	fmt.Printf("  %s           %s\n", output.Label("OAuth:"), oauthStatus)
	fmt.Printf("  %s           %s\n", output.Label("Token:"), tokenStatus)
	if scopeStr != "" {
		fmt.Printf("  %s          %s\n", output.Label("Scopes:"), scopeStr)
	}
	fmt.Println()
	return nil
}

func runAuthLogout(cmd *cobra.Command, _ []string) error {
	if err := client.Auth().Logout(); err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"status": "logged_out"})
	}

	fmt.Println()
	fmt.Printf("  %s Credentials removed.\n\n", output.Success("✓"))
	return nil
}

func runAuthSetAPIKey(cmd *cobra.Command, args []string) error {
	var key string
	if len(args) > 0 {
		key = args[0]
	} else {
		// Use hidden input for interactive mode
		fmt.Print("Enter API key: ")
		secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		key = string(secretBytes)
	}

	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if err := store.SaveAPIKey(key); err != nil {
		return fmt.Errorf("save API key: %w", err)
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"status": "api_key_saved"})
	}

	fmt.Println()
	fmt.Printf("  %s API key saved to keyring\n\n", output.Success("✓"))
	return nil
}

func runAuthSetClientSecret(cmd *cobra.Command, args []string) error {
	var secret string
	if len(args) > 0 {
		secret = args[0]
	} else {
		// Use hidden input for interactive mode
		fmt.Print("Enter client secret: ")
		secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read client secret: %w", err)
		}
		secret = string(secretBytes)
	}

	if secret == "" {
		return fmt.Errorf("client secret cannot be empty")
	}

	if err := store.SaveClientSecret(secret); err != nil {
		return fmt.Errorf("save client secret: %w", err)
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"status": "client_secret_saved"})
	}

	fmt.Println()
	fmt.Printf("  %s Client secret saved to keyring\n\n", output.Success("✓"))
	return nil
}
