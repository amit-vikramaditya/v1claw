package integrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Calendar ---

type mockCalendar struct {
	events []CalendarEvent
}

func (m *mockCalendar) Name() string { return "mock" }
func (m *mockCalendar) List(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
	var result []CalendarEvent
	for _, e := range m.events {
		if !e.Start.Before(start) && e.Start.Before(end) {
			result = append(result, e)
		}
	}
	return result, nil
}
func (m *mockCalendar) Get(ctx context.Context, id string) (*CalendarEvent, error) {
	for _, e := range m.events {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, nil
}
func (m *mockCalendar) Create(ctx context.Context, event CalendarEvent) (string, error) {
	event.ID = "new_event"
	m.events = append(m.events, event)
	return event.ID, nil
}
func (m *mockCalendar) Update(ctx context.Context, event CalendarEvent) error { return nil }
func (m *mockCalendar) Delete(ctx context.Context, id string) error           { return nil }

func TestCalendarManager_Today(t *testing.T) {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	cal := &mockCalendar{events: []CalendarEvent{
		{ID: "1", Summary: "Morning standup", Start: startOfToday.Add(9 * time.Hour), End: startOfToday.Add(10 * time.Hour)},
		{ID: "2", Summary: "Tomorrow meeting", Start: startOfToday.Add(25 * time.Hour), End: startOfToday.Add(26 * time.Hour)},
	}}

	mgr := NewCalendarManager(CalendarConfig{}, cal)
	events, err := mgr.Today(context.Background())
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "Morning standup", events[0].Summary)
}

func TestCalendarManager_NextEvent(t *testing.T) {
	now := time.Now()
	cal := &mockCalendar{events: []CalendarEvent{
		{ID: "1", Summary: "Next meeting", Start: now.Add(2 * time.Hour)},
	}}

	mgr := NewCalendarManager(CalendarConfig{}, cal)
	event, err := mgr.NextEvent(context.Background())
	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, "Next meeting", event.Summary)
}

func TestCalendarManager_NextEvent_None(t *testing.T) {
	cal := &mockCalendar{}
	mgr := NewCalendarManager(CalendarConfig{}, cal)
	event, err := mgr.NextEvent(context.Background())
	require.NoError(t, err)
	assert.Nil(t, event)
}

func TestCalendarManager_QuickAdd(t *testing.T) {
	cal := &mockCalendar{}
	mgr := NewCalendarManager(CalendarConfig{}, cal)
	id, err := mgr.QuickAdd(context.Background(), "Lunch", time.Now(), 60)
	require.NoError(t, err)
	assert.Equal(t, "new_event", id)
	assert.Len(t, cal.events, 1)
}

func TestFormatEvent(t *testing.T) {
	e := CalendarEvent{Summary: "Team sync", Start: time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 1, 15, 0, 0, 0, time.UTC)}
	result := FormatEvent(e)
	assert.Contains(t, result, "Team sync")
	assert.Contains(t, result, "📅")
}

func TestFormatEvent_AllDay(t *testing.T) {
	e := CalendarEvent{Summary: "Holiday", AllDay: true}
	result := FormatEvent(e)
	assert.Contains(t, result, "all day")
}

func TestFormatEventList_Empty(t *testing.T) {
	assert.Equal(t, "No events scheduled.", FormatEventList(nil))
}

// --- Mock Email ---

type mockEmail struct {
	messages []EmailMessage
}

func (m *mockEmail) Name() string { return "mock" }
func (m *mockEmail) Fetch(ctx context.Context, limit int) ([]EmailMessage, error) {
	if limit > len(m.messages) {
		limit = len(m.messages)
	}
	return m.messages[:limit], nil
}
func (m *mockEmail) Send(ctx context.Context, msg EmailMessage) error { return nil }
func (m *mockEmail) MarkRead(ctx context.Context, id string) error    { return nil }
func (m *mockEmail) UnreadCount(ctx context.Context) (int, error) {
	count := 0
	for _, msg := range m.messages {
		if !msg.Read {
			count++
		}
	}
	return count, nil
}

func TestEmailManager_Digest(t *testing.T) {
	email := &mockEmail{messages: []EmailMessage{
		{ID: "1", From: "alice@test.com", Subject: "Hello", Read: false},
		{ID: "2", From: "bob@test.com", Subject: "Meeting", Read: false},
	}}

	mgr := NewEmailManager(EmailConfig{}, email)
	digest, err := mgr.Digest(context.Background())
	require.NoError(t, err)
	assert.Contains(t, digest, "2 unread")
	assert.Contains(t, digest, "alice@test.com")
}

func TestEmailManager_Digest_Empty(t *testing.T) {
	email := &mockEmail{}
	mgr := NewEmailManager(EmailConfig{}, email)
	digest, err := mgr.Digest(context.Background())
	require.NoError(t, err)
	assert.Contains(t, digest, "No unread")
}

// --- Mock Smart Home ---

type mockSmartHome struct {
	devices []SmartHomeDevice
}

func (m *mockSmartHome) Name() string { return "mock" }
func (m *mockSmartHome) ListDevices(ctx context.Context) ([]SmartHomeDevice, error) {
	return m.devices, nil
}
func (m *mockSmartHome) GetDevice(ctx context.Context, id string) (*SmartHomeDevice, error) {
	for _, d := range m.devices {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, nil
}
func (m *mockSmartHome) SetState(ctx context.Context, id, state string) error {
	for i := range m.devices {
		if m.devices[i].ID == id {
			m.devices[i].State = state
			return nil
		}
	}
	return nil
}
func (m *mockSmartHome) CallService(ctx context.Context, domain, service string, data map[string]interface{}) error {
	return nil
}
func (m *mockSmartHome) Subscribe(ctx context.Context, handler func(SmartHomeDevice)) error {
	return nil
}

func TestSmartHomeManager_TurnOnOff(t *testing.T) {
	sh := &mockSmartHome{devices: []SmartHomeDevice{
		{ID: "light.living", Name: "Living Room Light", Type: "light", State: "off", Area: "living_room"},
	}}

	mgr := NewSmartHomeManager(SmartHomeConfig{}, sh)

	err := mgr.TurnOn(context.Background(), "light.living")
	require.NoError(t, err)
	assert.Equal(t, "on", sh.devices[0].State)

	err = mgr.TurnOff(context.Background(), "light.living")
	require.NoError(t, err)
	assert.Equal(t, "off", sh.devices[0].State)
}

func TestSmartHomeManager_ListByArea(t *testing.T) {
	sh := &mockSmartHome{devices: []SmartHomeDevice{
		{ID: "1", Area: "kitchen", Type: "light"},
		{ID: "2", Area: "bedroom", Type: "light"},
		{ID: "3", Area: "kitchen", Type: "sensor"},
	}}

	mgr := NewSmartHomeManager(SmartHomeConfig{}, sh)
	devices, err := mgr.ListByArea(context.Background(), "kitchen")
	require.NoError(t, err)
	assert.Len(t, devices, 2)
}

func TestSmartHomeManager_ListByType(t *testing.T) {
	sh := &mockSmartHome{devices: []SmartHomeDevice{
		{ID: "1", Type: "light", State: "on"},
		{ID: "2", Type: "sensor", State: "on"},
		{ID: "3", Type: "light", State: "off"},
	}}

	mgr := NewSmartHomeManager(SmartHomeConfig{}, sh)
	lights, err := mgr.ListByType(context.Background(), "light")
	require.NoError(t, err)
	assert.Len(t, lights, 2)
}

func TestSmartHomeManager_Status(t *testing.T) {
	sh := &mockSmartHome{devices: []SmartHomeDevice{
		{ID: "1", State: "on"},
		{ID: "2", State: "off"},
		{ID: "3", State: "on"},
	}}

	mgr := NewSmartHomeManager(SmartHomeConfig{}, sh)
	status, err := mgr.Status(context.Background())
	require.NoError(t, err)
	assert.Contains(t, status, "3 devices")
	assert.Contains(t, status, "2 on")
	assert.Contains(t, status, "1 off")
}
