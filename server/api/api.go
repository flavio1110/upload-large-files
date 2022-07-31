package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type apiServer struct {
	port       string
	httpServer *http.Server
	handler    http.Handler
	store      *memoryStore
}

func NewApiServer(port string) *apiServer {
	server := &apiServer{
		port:  port,
		store: NewStore(),
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
	router.Use(enforceJson)
	router.HandleFunc("/status", status).Methods(http.MethodGet)
	router.HandleFunc("/file/prepare", s.prepare).Methods(http.MethodPost)
	router.HandleFunc("/file/add-chunk/{id}", s.addChunk).Methods(http.MethodPost)
	router.HandleFunc("/file/finalize/{id}", s.finalize).Methods(http.MethodPost)
	router.HandleFunc("/file/download/{id}", s.download).Methods(http.MethodGet)
	s.handler = router
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (s *apiServer) prepare(w http.ResponseWriter, r *http.Request) {
	f, err := s.store.prepare()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error to prepeare", err)
		return
	}
	jsonRes(w, PrepareResponse{
		Id: f.id,
	})
}

func (s *apiServer) addChunk(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20)
	file, _, err := r.FormFile("chunk")
	if err != nil {
		fmt.Println("Error Retrieving the File", err)
		return
	}
	defer file.Close()
	id, err := getId(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}
	if err := s.store.addChunk(id, 1, file); err != nil {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("error to add chunk", err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *apiServer) finalize(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}

	if err := s.store.finalize(id); err != nil {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("error to finalize", err)
			return
		}
	}
}

func (s *apiServer) download(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}

	f, err := s.store.read(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error to download", err)
		return
	}
	defer f.Close()

	w.Header().Add("Content-Disposition", "attachment; filename=\"image.jpg\"")
	if _, err := io.Copy(w, f); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error to write response", err)
		return
	}
}

type PrepareResponse struct {
	Id uuid.UUID `json:"id"`
}

func jsonRes(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func enforceJson(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getId(r *http.Request) (uuid.UUID, error) {
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		return uuid.Nil, errors.New("id not provided")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("id not valid")
	}

	return uid, nil
}
