package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
	"github.com/qj0r9j0vc2/kko/internal/local"
	"github.com/spf13/cobra"
)

func usageError(cmd *cobra.Command, msg string) error {
	s := msg
	if ex := cmd.Example; ex != "" {
		s += "\n\n  Usage:\n" + ex
	}
	return fmt.Errorf("%s", s)
}

// FriendlyError extracts a user-friendly message from an error chain.
// It looks for *kakao.APIError and uses UserMessage(); otherwise returns the
// original text.
func FriendlyError(err error) string {
	var apiErr *kakao.APIError
	if errors.As(err, &apiErr) {
		return apiErr.UserMessage()
	}
	return err.Error()
}

func categoryList() string {
	keys := make([]string, 0, len(local.CategoryCodes))
	for k := range local.CategoryCodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
