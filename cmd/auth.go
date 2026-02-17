package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
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

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
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

	apiKeyStatus := "not configured"
	if cfg.APIKey != "" {
		masked := cfg.APIKey
		if len(masked) > 4 {
			masked = "..." + masked[len(masked)-4:]
		}
		apiKeyStatus = fmt.Sprintf("configured (KakaoAK %s)", masked)
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
			"api_key": apiKeyStatus,
			"oauth":   oauthStatus,
			"token":   tokenStatus,
			"scopes":  scopeStr,
		})
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n", output.Label("API Key:"), apiKeyStatus)
	fmt.Printf("  %s    %s\n", output.Label("OAuth:"), oauthStatus)
	fmt.Printf("  %s    %s\n", output.Label("Token:"), tokenStatus)
	if scopeStr != "" {
		fmt.Printf("  %s   %s\n", output.Label("Scopes:"), scopeStr)
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
