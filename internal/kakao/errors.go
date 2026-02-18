package kakao

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated: run `kko auth login`")
	ErrNoAPIKey         = errors.New("API key not configured: run `kko auth set-api-key`")
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Status  int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kakao api error %d: %s (HTTP %d)", e.Code, e.Message, e.Status)
}

var userMessages = map[int]string{
	-401: "API key is invalid. Run `kko auth set-api-key`",
	-2:   "Invalid parameter. Check your input.",
	-10:  "API limit exceeded. Try again later.",
	-530: "System error on Kakao's side. Try again later.",
	0:    "Permission denied. Check your API key has the required permissions.",
	-1:   "Invalid request. Check your input and try again.",
	-5:   "Endpoint not found. You may need to update kko.",
	-9:   "Not an allowed service. Enable the API in your Kakao app settings.",
}

func (e *APIError) UserMessage() string {
	if msg, ok := userMessages[e.Code]; ok {
		return msg
	}
	if e.Status == 401 || e.Status == 403 {
		return "Permission denied. Check your API key has the required permissions."
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("unexpected error (HTTP %d)", e.Status)
}

func ParseAPIError(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read response body", resp.StatusCode)
	}
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	apiErr.Status = resp.StatusCode
	return &apiErr
}
