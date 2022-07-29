package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type apiServer struct {
	port       string
	httpServer *http.Server
	handler    http.Handler
}

func NewApiServer(port string) *apiServer {
	server := &apiServer{
		port: port,
	}
	server.registerRoutes()
	return server
}

func (s *apiServer) Start() error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%s", s.port),
		Handler: s.handler,
	}

	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("start api server: %w", err)
	}
	return nil
}

func (s *apiServer) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("stop api server: %w", err)
	}
	return nil
}

func (s *apiServer) registerRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/status", status)
	s.handler = router
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
