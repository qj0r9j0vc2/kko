package calendar

import "time"

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Time        EventTime `json:"time"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	CalendarID  string    `json:"calendar_id,omitempty"`
}

type EventTime struct {
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	TimeZone string    `json:"time_zone"`
	AllDay   bool      `json:"all_day"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type CreateEventRequest struct {
	Title       string     `json:"title"`
	Time        *EventTime `json:"time"`
	Description string     `json:"description,omitempty"`
	Location    string     `json:"location,omitempty"`
}

type Todo struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	DueAt     time.Time `json:"due_at,omitempty"`
	Completed bool      `json:"completed"`
}

type TodosResponse struct {
	Todos []Todo `json:"todos"`
}
