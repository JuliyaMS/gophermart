package server

import (
	"github.com/JuliyaMS/gophermart/internal/config"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	r      *Router
	logger *zap.SugaredLogger
}

func NewServer(log *zap.SugaredLogger, ro *Router) *Server {

	return &Server{
		logger: log,
		r:      ro,
	}
}

func (s *Server) Start() {

	s.logger.Infow("Get router")
	router := s.r.GetRouter()

	s.logger.Infow("Start gophermart", "address", config.RunServerURL)
	if err := http.ListenAndServe(config.RunServerURL, router); err != nil {
		s.logger.Fatalf(err.Error(), "event", "start server")
		return

	}
}
