package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	Logger *logrus.Logger
	Port   string
}

// Handler builds the routed, logged handler so tests can exercise it without listening.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	return LoggingMiddleware(s.Logger)(mux)
}

// Serve starts the HTTP server until it exits.
func (s *Server) Serve() {
	server := http.Server{
		Addr:              ":" + s.Port,
		Handler:           s.Handler(),
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.Logger.WithField("port", s.Port).Info("Server.Serve.listening")
	if err := server.ListenAndServe(); err != nil {
		s.Logger.WithError(err).Error("Server.Serve.listen error")
	}
	s.Logger.Info("Server.Serve.shutting down")
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
