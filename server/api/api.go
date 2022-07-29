package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
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
	router.HandleFunc("/status", status).Methods(http.MethodGet)
	router.HandleFunc("/file/download", status).Methods(http.MethodGet)
	router.HandleFunc("/file/prepare", status).Methods(http.MethodPost)
	router.HandleFunc("/file/upload-chunk", status).Methods(http.MethodPost)
	router.HandleFunc("/file/finalize", status).Methods(http.MethodPost)
	s.handler = router
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func prepare(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req PrepareRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := PrepareResponse{
		id: uuid.New(),
	}

	JsonResponse(w, res)
}

func upload(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func finalize(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

type PrepareRequest struct {
	numerOfChunks int
	sizeInBytes   int64
}

type PrepareResponse struct {
	id uuid.UUID
}

func JsonResponse(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
