package local

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
)

const (
	keywordURL  = "https://dapi.kakao.com/v2/local/search/keyword.json"
	categoryURL = "https://dapi.kakao.com/v2/local/search/category.json"
	addressURL  = "https://dapi.kakao.com/v2/local/search/address.json"
)

type Service struct {
	client *kakao.Client
}

func NewService(client *kakao.Client) *Service {
	return &Service{client: client}
}

type SearchOptions struct {
	Query    string
	Category string
	X        string
	Y        string
	Radius   int
	Sort     string
	Page     int
	Size     int
}

func (s *Service) KeywordSearch(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	params := url.Values{}
	params.Set("query", opts.Query)
	if opts.X != "" && opts.Y != "" {
		params.Set("x", opts.X)
		params.Set("y", opts.Y)
	}
	if opts.Radius > 0 {
		params.Set("radius", strconv.Itoa(opts.Radius))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Size > 0 {
		params.Set("size", strconv.Itoa(opts.Size))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	body, err := s.client.Get(ctx, keywordURL, params)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse keyword search response: %w", err)
	}
	return &result, nil
}

func (s *Service) CategorySearch(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	code, ok := CategoryCodes[opts.Category]
	if !ok {
		code = opts.Category
	}

	params := url.Values{}
	params.Set("category_group_code", code)
	if opts.X != "" && opts.Y != "" {
		params.Set("x", opts.X)
		params.Set("y", opts.Y)
	}
	if opts.Radius > 0 {
		params.Set("radius", strconv.Itoa(opts.Radius))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Size > 0 {
		params.Set("size", strconv.Itoa(opts.Size))
	}

	body, err := s.client.Get(ctx, categoryURL, params)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse category search response: %w", err)
	}
	return &result, nil
}

func (s *Service) AddressSearch(ctx context.Context, query string) (*AddressResult, error) {
	params := url.Values{}
	params.Set("query", query)

	body, err := s.client.Get(ctx, addressURL, params)
	if err != nil {
		return nil, err
	}

	var result AddressResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse address search response: %w", err)
	}
	return &result, nil
}

func (s *Service) ResolveLocation(ctx context.Context, location string) (x, y string, err error) {
	if coords := parseCoords(location); coords != nil {
		return coords[0], coords[1], nil
	}

	addrResult, addrErr := s.AddressSearch(ctx, location)
	if addrErr == nil && len(addrResult.Documents) > 0 {
		return addrResult.Documents[0].X, addrResult.Documents[0].Y, nil
	}

	kwResult, err := s.KeywordSearch(ctx, SearchOptions{Query: location, Size: 1})
	if err != nil {
		return "", "", fmt.Errorf("resolve location %q: %w", location, err)
	}
	if len(kwResult.Documents) == 0 {
		return "", "", fmt.Errorf("location not found: %q", location)
	}
	return kwResult.Documents[0].X, kwResult.Documents[0].Y, nil
}

func parseCoords(s string) []string {
	for _, sep := range []string{","} {
		parts := strings.SplitN(s, sep, 2)
		if len(parts) == 2 {
			p0 := strings.TrimSpace(parts[0])
			p1 := strings.TrimSpace(parts[1])
			_, err1 := strconv.ParseFloat(p0, 64)
			_, err2 := strconv.ParseFloat(p1, 64)
			if err1 == nil && err2 == nil {
				return []string{p0, p1}
			}
		}
	}
	return nil
}
