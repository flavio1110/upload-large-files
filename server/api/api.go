package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

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
		store: NewStore("temp/"),
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
	router.HandleFunc("/file/add-chunk/{id}/{number}", s.addChunk).Methods(http.MethodPost)
	router.HandleFunc("/file/finalize/{id}", s.finalize).Methods(http.MethodPost)
	router.HandleFunc("/file/download/{id}", s.download).Methods(http.MethodGet)
	s.handler = router
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (s *apiServer) prepare(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	request := struct {
		Name        string `json:"name"`
		ContentType string `json:"content_type"`
	}{}

	if err := json.Unmarshal(reqBody, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := s.store.prepare(request.Name, request.ContentType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error to prepeare", err)
		return
	}
	fmt.Println("prepared", f.id, request.Name, request.ContentType)
	jsonRes(w, PrepareResponse{
		Id: f.id,
	})
}

func (s *apiServer) addChunk(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20)
	file, _, err := r.FormFile("chunk")
	if err != nil {
		log.Println("error to read chunk", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	id, err := getId(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}

	number, err := getNumber(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}

	if err := s.store.addChunk(id, number, file); err != nil {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("error to add chunk", err)
			return
		}
	}
	fmt.Println("Added chunk", id, number)
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
	fmt.Println("finalized", id)
}

func (s *apiServer) download(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//TODO: return error details
		return
	}

	file, reader, err := s.store.read(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error to download", err)
		return
	}
	defer reader.Close()

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.name))
	w.Header().Del("Content-Type")
	w.Header().Add("Content-Type", file.contentType)
	if _, err := io.Copy(w, reader); err != nil {
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

func getNumber(r *http.Request) (int, error) {
	params := mux.Vars(r)
	param, ok := params["number"]
	if !ok {
		return 0, errors.New("number not provided")
	}

	number, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.New("id not valid")
	}

	return number, nil
}
