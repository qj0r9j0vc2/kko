package mobility

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
)

const (
	directionsURL       = "https://apis-navi.kakaomobility.com/v1/directions"
	futureDirectionsURL = "https://apis-navi.kakaomobility.com/v1/future/directions"
)

type Service struct {
	client *kakao.Client
}

func NewService(client *kakao.Client) *Service {
	return &Service{client: client}
}

type DirectionsOptions struct {
	OriginX    string
	OriginY    string
	DestX      string
	DestY      string
	Priority   string
	Avoid      string
	DepartTime string
}

func (s *Service) Directions(ctx context.Context, opts DirectionsOptions) (*DirectionsResult, error) {
	endpoint := directionsURL
	if opts.DepartTime != "" {
		endpoint = futureDirectionsURL
	}

	params := url.Values{}
	params.Set("origin", opts.OriginX+","+opts.OriginY)
	params.Set("destination", opts.DestX+","+opts.DestY)

	if opts.Priority != "" {
		params.Set("priority", opts.Priority)
	}
	if opts.Avoid != "" {
		params.Set("avoid", opts.Avoid)
	}
	if opts.DepartTime != "" {
		params.Set("departure_time", opts.DepartTime)
	}

	body, err := s.client.Get(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var result DirectionsResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse directions response: %w", err)
	}
	return &result, nil
}
