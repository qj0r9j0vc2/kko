package cmd

import (
	"fmt"
	"strings"

	"github.com/qj0r9j0vc2/kko/internal/message"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var msgCmd = &cobra.Command{
	Use:   "msg [message]",
	Short: "Send a KakaoTalk message",
	Long:  "Send a message via KakaoTalk to yourself or a friend. Requires OAuth login.",
	Example: `  kko msg "heading to the office now"
  kko msg "meeting at 3pm" --link "https://zoom.us/j/123"
  kko msg "late 10 min" --to friend_uuid`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMsg,
}

var (
	msgTo   string
	msgLink string
)

func init() {
	rootCmd.AddCommand(msgCmd)

	msgCmd.Flags().StringVarP(&msgTo, "to", "t", "me", "recipient: 'me' or friend UUID")
	msgCmd.Flags().StringVarP(&msgLink, "link", "l", "", "attach a URL to the message")
}

func runMsg(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc := message.NewService(client)

	text := strings.Join(args, " ")

	var template *message.TemplateObject
	if msgLink != "" {
		template = message.BuildLinkTemplate(text, msgLink)
	} else {
		template = message.BuildTextTemplate(text)
	}

	if cfg.Output.Format == "json" {
		result := map[string]interface{}{
			"to":      msgTo,
			"message": text,
			"link":    msgLink,
		}
		var sendErr error
		if msgTo == "me" || msgTo == "" {
			sendErr = svc.SendToMe(ctx, template)
		} else {
			sendErr = svc.SendToFriend(ctx, msgTo, template)
		}
		if sendErr != nil {
			result["status"] = "error"
			result["error"] = sendErr.Error()
		} else {
			result["status"] = "sent"
		}
		return output.PrintJSON(result)
	}

	if msgTo == "me" || msgTo == "" {
		if err := svc.SendToMe(ctx, template); err != nil {
			return err
		}
		msg := "Message sent to myself via KakaoTalk."
		if msgLink != "" {
			msg += " (with link)"
		}
		fmt.Println()
		fmt.Printf("  %s %s\n\n", output.Success("✓"), msg)
	} else {
		if err := svc.SendToFriend(ctx, msgTo, template); err != nil {
			return err
		}
		fmt.Println()
		fmt.Printf("  %s Message sent to friend via KakaoTalk.\n\n", output.Success("✓"))
	}

	return nil
}
