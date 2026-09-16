package internalhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
)

// type mockLogger struct{}
// func (ml *mockLogger) log(level string, msg string) {}
// func (t *mockLogger) Debug(f string, args ...interface{})  {}
// func (t *mockLogger) Info(f string, args ...interface{})  {}
// func (t *mockLogger) Warning(f string, args ...interface{})  {}
// func (t *mockLogger) Error(f string, args ...interface{}) {}

type mockStorage struct {
	events []storage.Event
}

func (m *mockStorage) CreateEvent(e *storage.Event) error {
	e.ID = int64(len(m.events)) + 1
	m.events = append(m.events, *e)
	return nil
}
func (m *mockStorage) UpdateEvent(e *storage.Event) error { return nil }
func (m *mockStorage) DeleteEvent(id int64) error         { return nil }
func (m *mockStorage) ListEvents() ([]storage.Event, error) {
	return m.events, nil
}

func TestHTTP_CreateEvent(t *testing.T) {
	mockLogger := logger.New("DEBUG")
	mockStorage := &mockStorage{}
	calendar := app.New(mockLogger, mockStorage)
	server := NewServer(ServerConf{Host: "127.0.0.1", Port: "8080"}, calendar, mockLogger)

	eventReq := storage.Event{
		Title:       "Тестовый созвон",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Description: "Юнит-тестирование",
		User:        "test_user",
	}
	fmt.Print(eventReq)
	body, _ := json.Marshal(eventReq)

	req, err := http.NewRequest(http.MethodPost, "/create_event", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var responceEvent storage.Event
	if err := json.NewDecoder(rr.Body).Decode(&responceEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// fmt.Print(responceEvent)
	if responceEvent.ID != 1 {
		t.Errorf("expected generated event ID to be '1', got '%d'", responceEvent.ID)
	}
}

func TestHTTP_ListEventsForDay(t *testing.T) {
	mockLogger := logger.New("DEBUG")
	mockStorage := &mockStorage{}
	targetTime := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mockStorage.Add(&storage.Event{
		Title:     "Событие №1",
		StartTime: targetTime,
		User:      "test-user",
	})

	calendar := app.New(mockLogger, mockStorage)
	server := NewServer(ServerConf{Host: "127.0.0.1", Port: "8080"}, calendar, mockLogger)

	req, err := http.NewRequest(http.MethodGet, "/events_for_day?day=2026-09-16", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string][]storage.Event
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp["events"]) != 1 {
		t.Errorf("expected 1 event in list, got %d", len(resp["events"]))
	}
}
