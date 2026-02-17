package kakao

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{Code: -401, Message: "Unauthorized", Status: 401}
	got := err.Error()
	want := "kakao api error -401: Unauthorized (HTTP 401)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAPIError_UserMessage(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		want string
	}{
		{-401, "Unauthorized", "API key is invalid. Run `kko config set api_key <KEY>`"},
		{-2, "Bad Request", "Invalid parameter. Check your input."},
		{-10, "TooManyRequests", "API limit exceeded. Try again later."},
		{-530, "Internal Error", "System error on Kakao's side. Try again later."},
		{-999, "Unknown error", "Unknown error"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			err := &APIError{Code: tt.code, Message: tt.msg}
			if got := err.UserMessage(); got != tt.want {
				t.Errorf("UserMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
