package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
)

const (
	baseURL = "https://dapi.kakao.com/v2/search"
)

type Service struct {
	client *kakao.Client
}

func NewService(client *kakao.Client) *Service {
	return &Service{client: client}
}

type SearchOptions struct {
	Query string
	Type  string
	Sort  string
	Page  int
	Size  int
}

func (s *Service) Search(ctx context.Context, opts SearchOptions) (*WebSearchResponse, error) {
	searchType := opts.Type
	if searchType == "" {
		searchType = "web"
	}

	endpoint := fmt.Sprintf("%s/%s", baseURL, searchType)
	params := url.Values{}
	params.Set("query", opts.Query)
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Size > 0 {
		params.Set("size", strconv.Itoa(opts.Size))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	body, err := s.client.Get(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var result WebSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}
	return &result, nil
}

func (s *Service) SearchBlog(ctx context.Context, opts SearchOptions) (*BlogSearchResponse, error) {
	endpoint := fmt.Sprintf("%s/blog", baseURL)
	params := url.Values{}
	params.Set("query", opts.Query)
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Size > 0 {
		params.Set("size", strconv.Itoa(opts.Size))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	body, err := s.client.Get(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var result BlogSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse blog search response: %w", err)
	}
	return &result, nil
}

func (s *Service) SearchCafe(ctx context.Context, opts SearchOptions) (*CafeSearchResponse, error) {
	endpoint := fmt.Sprintf("%s/cafe", baseURL)
	params := url.Values{}
	params.Set("query", opts.Query)
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Size > 0 {
		params.Set("size", strconv.Itoa(opts.Size))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	body, err := s.client.Get(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var result CafeSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse cafe search response: %w", err)
	}
	return &result, nil
}

func StripHTML(s string) string {
	var result []byte
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
			continue
		}
		if s[i] == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, s[i])
		}
	}
	return string(result)
}
