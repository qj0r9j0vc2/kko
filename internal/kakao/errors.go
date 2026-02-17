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
	ErrNoAPIKey         = errors.New("API key not configured: run `kko config set api_key <KEY>`")
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
	-401: "API key is invalid. Run `kko config set api_key <KEY>`",
	-2:   "Invalid parameter. Check your input.",
	-10:  "API limit exceeded. Try again later.",
	-530: "System error on Kakao's side. Try again later.",
}

func (e *APIError) UserMessage() string {
	if msg, ok := userMessages[e.Code]; ok {
		return msg
	}
	return e.Message
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
