package api

import (
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type item struct {
	id          uuid.UUID
	tempPath    string
	chunckPaths []string
	closed      bool
	finalPath   string
}

type memoryStore struct {
	files map[uuid.UUID]item
}

func NewStore() *memoryStore {
	return &memoryStore{
		files: make(map[uuid.UUID]item),
	}
}

func (s *memoryStore) prepare() (item, error) {
	id := uuid.New()
	err := os.Mkdir(id.String(), os.ModePerm)
	if err != nil {
		return item{}, fmt.Errorf("create temp directory: %w", err)
	}

	i := item{
		id:       uuid.New(),
		tempPath: id.String(),
		closed:   false,
	}
	return i, nil
}

func (s *memoryStore) addChunk(id uuid.UUID, r io.Reader) error {
	i, ok := s.files[id]
	if !ok {
		return fmt.Errorf("file not found with id %q", id)
	}

	if i.closed {
		return fmt.Errorf("file %q is already closed", id)
	}

	w, err := os.CreateTemp(i.tempPath, "*")
	if err != nil {
		return fmt.Errorf("create chunk file: %w", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("copy contents to temp file: %w", err)
	}
	i.chunckPaths = append(i.chunckPaths, w.Name())
	return nil
}

func (s *memoryStore) finalize(id uuid.UUID, r io.Reader) (item, error) {
	i, ok := s.files[id]
	if !ok {
		return item{}, fmt.Errorf("file not found with id %q", id)
	}

	if i.closed {
		return item{}, fmt.Errorf("file %q is already closed", id)
	}

	w, err := os.CreateTemp(i.tempPath, fmt.Sprintf("%s_FINAL", i.id))

	if err != nil {
		return item{}, fmt.Errorf("create final file: %w", err)
	}
	defer w.Close()

	for _, path := range i.chunckPaths {
		r, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return item{}, fmt.Errorf("read temp file %q: %w", path, err)
		}
		defer r.Close()
		if _, err := io.Copy(w, r); err != nil {
			return item{}, fmt.Errorf("copy contents from temp file %q: %w", path, err)
		}
	}

	defer func() {
		for _, path := range i.chunckPaths {
			os.Remove(path)
		}
	}()

	if _, err := io.Copy(w, r); err != nil {
		return item{}, fmt.Errorf("copy contents to temp file: %w", err)
	}
	i.chunckPaths = append(i.chunckPaths, w.Name())
	return i, nil
}
