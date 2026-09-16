package internalhttp

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
)

type Server struct {
	calendarApp *app.App
	logger      *logger.Logger
	httpServer  *http.Server
	mux         *http.ServeMux
}

type ServerConf struct {
	Host string
	Port string
}

type Application interface{}

type Storage interface{}

func NewServer(serverConf ServerConf, app *app.App /*storage Storage,*/, logger *logger.Logger) *Server {
	server := &Server{
		calendarApp: app,
		logger:      logger,
	}
	mux := http.NewServeMux()

	mux.Handle("/create_event", server.loggingMiddleware(http.HandlerFunc(server.handleCreateEvent)))
	mux.Handle("/update_event", server.loggingMiddleware(http.HandlerFunc(server.handleUpdateEvent)))
	mux.Handle("/delete_event", server.loggingMiddleware(http.HandlerFunc(server.handleDeleteEvent)))
	mux.Handle("/events_for_day", server.loggingMiddleware(http.HandlerFunc(server.handleEventsForDay)))
	mux.Handle("/events_for_week", server.loggingMiddleware(http.HandlerFunc(server.handleEventsForWeek)))
	mux.Handle("/events_for_month", server.loggingMiddleware(http.HandlerFunc(server.handleEventsForMonth)))

	mux.Handle("/hello", server.loggingMiddleware(http.HandlerFunc(server.helloHandler)))
	mux.Handle("/", server.loggingMiddleware(http.HandlerFunc(server.helloHandler)))
	server.mux = mux

	server.httpServer = &http.Server{
		Addr:              net.JoinHostPort(serverConf.Host, serverConf.Port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	return server
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server on %s", s.httpServer.Addr)

	errorChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	select {
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, world!"))
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req storage.Event
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request JSON", http.StatusBadRequest)
		return
	}

	dbEvent := storage.Event{
		Title:       req.Title,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Description: req.Description,
		User:        req.User,
	}

	if err := s.calendarApp.CreateEvent(r.Context(), &dbEvent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s.toHTTPModel(dbEvent))
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid or missing id parameter", http.StatusBadRequest)
		return
	}

	var req storage.Event
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request JSON", http.StatusBadRequest)
		return
	}

	dbEvent := storage.Event{
		ID:          id,
		Title:       req.Title,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Description: req.Description,
		User:        req.User,
	}

	if err := s.calendarApp.UpdateEvent(r.Context(), &dbEvent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.toHTTPModel(dbEvent))
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if err := s.calendarApp.DeleteEvent(r.Context(), idStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleEventsForDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dayStr := r.URL.Query().Get("day")
	targetTime, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := s.calendarApp.ListEventsForDay(r.Context(), targetTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"events": s.toHTTPModelSlice(events)})
}

func (s *Server) handleEventsForWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	weekStartStr := r.URL.Query().Get("week_start")
	targetTime, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := s.calendarApp.ListEventsForWeek(r.Context(), targetTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"events": s.toHTTPModelSlice(events)})
}

func (s *Server) handleEventsForMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	monthStartStr := r.URL.Query().Get("month_start")
	targetTime, err := time.Parse("2006-01-02", monthStartStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := s.calendarApp.ListEventsForMonth(r.Context(), targetTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"events": s.toHTTPModelSlice(events)})
}

func (s *Server) toHTTPModel(e storage.Event) storage.Event {
	return storage.Event{
		ID:          e.ID,
		Title:       e.Title,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
		Description: e.Description,
		User:        e.User,
	}
}

func (s *Server) toHTTPModelSlice(events []storage.Event) []storage.Event {
	res := make([]storage.Event, len(events))
	for i, e := range events {
		res[i] = s.toHTTPModel(e)
	}
	return res
}
