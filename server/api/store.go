package api

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/google/uuid"
)

type item struct {
	id          uuid.UUID
	tempPath    string
	chunckPaths []string
	closed      bool
	finalPath   string
	name        string
	contentType string
}

type memoryStore struct {
	files    map[uuid.UUID]item
	basePath string
}

func NewStore(basePath string) *memoryStore {
	return &memoryStore{
		files:    make(map[uuid.UUID]item),
		basePath: basePath,
	}
}

func (s *memoryStore) prepare(name string, contentType string) (item, error) {
	id := uuid.New()
	temp := s.basePath + id.String()
	err := os.Mkdir(temp, os.ModePerm)
	if err != nil {
		return item{}, fmt.Errorf("create temp directory: %w", err)
	}

	i := item{
		id:          id,
		tempPath:    temp,
		closed:      false,
		name:        name,
		contentType: contentType,
	}
	s.files[i.id] = i
	return i, nil
}

func (s *memoryStore) addChunk(id uuid.UUID, number int, r io.Reader) error {
	i, ok := s.files[id]
	if !ok {
		return fmt.Errorf("file not found with id %q", id)
	}

	if i.closed {
		return fmt.Errorf("file %q is already closed", id)
	}

	w, err := os.Create(fmt.Sprintf("%s/%d", i.tempPath, number))
	if err != nil {
		return fmt.Errorf("create chunk file: %w", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("copy contents to temp file: %w", err)
	}
	i.chunckPaths = append(i.chunckPaths, w.Name())
	s.files[id] = i
	return nil
}

func (s *memoryStore) finalize(id uuid.UUID) error {
	i, ok := s.files[id]
	if !ok {
		return fmt.Errorf("file not found with id %q", id)
	}

	if i.closed {
		return fmt.Errorf("file %q is already closed", id)
	}

	w, err := os.CreateTemp(i.tempPath, fmt.Sprintf("%s_FINAL", i.id))

	if err != nil {
		return fmt.Errorf("create final file: %w", err)
	}
	defer w.Close()
	sort.Strings(i.chunckPaths)
	for _, path := range i.chunckPaths {
		r, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return fmt.Errorf("read temp file %q: %w", path, err)
		}
		defer r.Close()
		if _, err := io.Copy(w, r); err != nil {
			return fmt.Errorf("copy contents from temp file %q: %w", path, err)
		}
	}

	i.closed = true
	i.finalPath = w.Name()
	s.files[id] = i

	return nil
}

func (s *memoryStore) read(id uuid.UUID) (item, io.ReadCloser, error) {
	i, ok := s.files[id]
	if !ok {
		return item{}, nil, fmt.Errorf("file not found with id %q", id)
	}

	if !i.closed {
		return item{}, nil, fmt.Errorf("file %q is not yet closed", id)
	}

	f, err := os.OpenFile(i.finalPath, os.O_RDONLY, 0644)
	if err != nil {
		return item{}, nil, fmt.Errorf("read final file: %w", err)
	}

	return i, f, nil
}
