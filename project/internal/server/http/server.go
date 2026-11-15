package http

import (
	"encoding/json"
	"net/http"

	"github.com/TheTeaToper/OTUS/project/internal/app"
)

type Server struct {
	service *app.AntibruteforceService
	mux     *http.ServeMux
}

func NewServer(service *app.AntibruteforceService) *Server {
	server := &Server{
		service: service,
		mux:     http.NewServeMux(),
	}
	server.mux.HandleFunc("POST /api/v1/blacklist", server.handleAddToBlacklist)
	server.mux.HandleFunc("DELETE /api/v1/blacklist", server.handleRemoveFromBlacklist)

	server.mux.HandleFunc("POST /api/v1/whitelist", server.handleAddToWhitelist)
	server.mux.HandleFunc("DELETE /api/v1/whitelist", server.handleRemoveFromWhitelist)

	server.mux.HandleFunc("POST /api/v1/auth/validate", server.handleValidateAuthorization)
	server.mux.HandleFunc("POST /api/v1/buckets/reset", server.handleResetBuckets)

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleAddToBlacklist(w http.ResponseWriter, r *http.Request) {
	s.modifyList(w, r, s.service.AddToBlacklist)
}

func (s *Server) handleRemoveFromBlacklist(w http.ResponseWriter, r *http.Request) {
	s.modifyList(w, r, s.service.RemoveFromBlacklist)
}

func (s *Server) handleAddToWhitelist(w http.ResponseWriter, r *http.Request) {
	s.modifyList(w, r, s.service.AddToWhitelist)
}

func (s *Server) handleRemoveFromWhitelist(w http.ResponseWriter, r *http.Request) {
	s.modifyList(w, r, s.service.RemoveFromWhitelist)
}

func (s *Server) handleValidateAuthorization(w http.ResponseWriter, r *http.Request) {
	var request AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.respondWithError(w, "invalid json body")
		return
	}

	if request.Login == "" || request.Password == "" || request.IP == "" {
		s.respondWithError(w, "login, password and ip are required")
		return
	}

	ok := s.service.ValidateAuthorization(request.Login, request.Password, request.IP)
	s.respondWithJSON(w, http.StatusOK, AuthResponse{OK: ok})
}

func (s *Server) handleResetBuckets(w http.ResponseWriter, r *http.Request) {
	var req ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, "invalid json body")
		return
	}

	if req.Login == "" || req.IP == "" {
		s.respondWithError(w, "login and ip are required")
		return
	}

	s.service.ResetBuckets(req.Login, req.IP)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) modifyList(w http.ResponseWriter, r *http.Request, action func(string) error) {
	var req CidrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, "invalid json body")
		return
	}

	if err := action(req.Subnet); err != nil {
		s.respondWithError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) respondWithError(w http.ResponseWriter, msg string) {
	s.respondWithJSON(w, http.StatusBadRequest, ErrorResponse{Error: msg})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
