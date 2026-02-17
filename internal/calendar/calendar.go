package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/qj0r9j0vc2/kko/internal/kakao"
)

const (
	eventsURL      = "https://kapi.kakao.com/v2/api/calendar/events"
	createEventURL = "https://kapi.kakao.com/v2/api/calendar/create/event"
	deleteEventURL = "https://kapi.kakao.com/v2/api/calendar/delete/event"
	todosURL       = "https://kapi.kakao.com/v2/api/calendar/tasks"
)

type Service struct {
	client *kakao.Client
}

func NewService(client *kakao.Client) *Service {
	return &Service{client: client}
}

func (s *Service) ListEvents(ctx context.Context, from, to time.Time) ([]Event, error) {
	params := url.Values{}
	params.Set("from", from.Format(time.RFC3339))
	params.Set("to", to.Format(time.RFC3339))

	body, err := s.client.Get(ctx, eventsURL, params, kakao.WithOAuth())
	if err != nil {
		return nil, err
	}

	var resp EventsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse events: %w", err)
	}
	return resp.Events, nil
}

func (s *Service) CreateEvent(ctx context.Context, req *CreateEventRequest) (*Event, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	form := url.Values{}
	form.Set("event", string(reqJSON))

	body, err := s.client.Post(ctx, createEventURL, form, kakao.WithOAuth())
	if err != nil {
		return nil, err
	}

	var event Event
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("parse created event: %w", err)
	}
	return &event, nil
}

func (s *Service) DeleteEvent(ctx context.Context, eventID string) error {
	params := url.Values{}
	params.Set("event_id", eventID)

	_, err := s.client.Delete(ctx, deleteEventURL, params, kakao.WithOAuth())
	return err
}

func (s *Service) ListTodos(ctx context.Context) ([]Todo, error) {
	body, err := s.client.Get(ctx, todosURL, nil, kakao.WithOAuth())
	if err != nil {
		return nil, err
	}

	var resp TodosResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return resp.Todos, nil
}

func (s *Service) CreateTodo(ctx context.Context, content string, dueAt *time.Time) (*Todo, error) {
	form := url.Values{}
	form.Set("content", content)
	if dueAt != nil {
		form.Set("due_at", dueAt.Format(time.RFC3339))
	}

	body, err := s.client.Post(ctx, todosURL, form, kakao.WithOAuth())
	if err != nil {
		return nil, err
	}

	var todo Todo
	if err := json.Unmarshal(body, &todo); err != nil {
		return nil, fmt.Errorf("parse created todo: %w", err)
	}
	return &todo, nil
}
