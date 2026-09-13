package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     Logger
}

type ServerConf struct {
	Host string
	Port string
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
}

type Application interface{}

type Storage interface{}

func NewServer(serverConf ServerConf, app Application, storage Storage, logger Logger) *Server {
	server := &Server{
		logger: logger,
	}
	mux := http.NewServeMux()
	mux.Handle("/hello", server.loggingMiddleware(http.HandlerFunc(server.helloHandler)))
	mux.Handle("/", server.loggingMiddleware(http.HandlerFunc(server.helloHandler)))

	server.httpServer = &http.Server{
		Addr:              net.JoinHostPort(serverConf.Host, serverConf.Port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	return server
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("Starting HTTP server on %s", s.httpServer.Addr))

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
