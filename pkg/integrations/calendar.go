package integrations

import (
	"context"
	"fmt"
	"time"
)

// CalendarEvent represents a calendar event.
type CalendarEvent struct {
	ID          string            `json:"id"`
	Summary     string            `json:"summary"`
	Description string            `json:"description,omitempty"`
	Location    string            `json:"location,omitempty"`
	Start       time.Time         `json:"start"`
	End         time.Time         `json:"end"`
	AllDay      bool              `json:"all_day,omitempty"`
	Recurrence  string            `json:"recurrence,omitempty"`
	Reminders   []time.Duration   `json:"reminders,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CalendarProvider is the interface for calendar backends.
type CalendarProvider interface {
	// List returns events in the given time range.
	List(ctx context.Context, start, end time.Time) ([]CalendarEvent, error)
	// Get returns a specific event by ID.
	Get(ctx context.Context, id string) (*CalendarEvent, error)
	// Create creates a new event.
	Create(ctx context.Context, event CalendarEvent) (string, error)
	// Update modifies an existing event.
	Update(ctx context.Context, event CalendarEvent) error
	// Delete removes an event.
	Delete(ctx context.Context, id string) error
	// Name returns the provider name.
	Name() string
}

// CalendarConfig holds calendar integration configuration.
type CalendarConfig struct {
	Enabled     bool   `json:"enabled"`
	Provider    string `json:"provider"`     // "caldav", "google", "ical"
	URL         string `json:"url"`          // CalDAV URL or iCal feed URL
	Username    string `json:"username"`
	Password    string `json:"password"`
	CalendarID  string `json:"calendar_id"`  // Specific calendar
	ReminderMin int    `json:"reminder_min"` // Default reminder minutes
}

// CalendarManager provides a unified interface over calendar providers
// with proactive reminder support.
type CalendarManager struct {
	provider CalendarProvider
	config   CalendarConfig
}

// NewCalendarManager creates a calendar manager.
func NewCalendarManager(cfg CalendarConfig, provider CalendarProvider) *CalendarManager {
	return &CalendarManager{config: cfg, provider: provider}
}

// Today returns today's events.
func (m *CalendarManager) Today(ctx context.Context) ([]CalendarEvent, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return m.provider.List(ctx, start, end)
}

// Upcoming returns events in the next N hours.
func (m *CalendarManager) Upcoming(ctx context.Context, hours int) ([]CalendarEvent, error) {
	now := time.Now()
	return m.provider.List(ctx, now, now.Add(time.Duration(hours)*time.Hour))
}

// NextEvent returns the next upcoming event.
func (m *CalendarManager) NextEvent(ctx context.Context) (*CalendarEvent, error) {
	events, err := m.Upcoming(ctx, 24)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	return &events[0], nil
}

// QuickAdd creates an event from minimal information.
func (m *CalendarManager) QuickAdd(ctx context.Context, summary string, start time.Time, durationMin int) (string, error) {
	event := CalendarEvent{
		Summary: summary,
		Start:   start,
		End:     start.Add(time.Duration(durationMin) * time.Minute),
	}
	return m.provider.Create(ctx, event)
}

// FormatEvent returns a human-readable string for an event.
func FormatEvent(e CalendarEvent) string {
	if e.AllDay {
		return fmt.Sprintf("📅 %s (all day)", e.Summary)
	}
	return fmt.Sprintf("📅 %s (%s - %s)",
		e.Summary,
		e.Start.Format("3:04 PM"),
		e.End.Format("3:04 PM"),
	)
}

// FormatEventList formats multiple events.
func FormatEventList(events []CalendarEvent) string {
	if len(events) == 0 {
		return "No events scheduled."
	}
	result := ""
	for _, e := range events {
		result += FormatEvent(e) + "\n"
	}
	return result
}
