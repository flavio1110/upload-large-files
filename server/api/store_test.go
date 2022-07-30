package api

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreFlow(t *testing.T) {
	sut := NewStore()
	var file item

	t.Run("test prepare", func(t *testing.T) {
		var err error
		file, err = sut.prepare()
		assert := assert.New(t)
		assert.NoError(err)
		i, ok := sut.files[file.id]
		assert.True(ok)
		assert.Equal(i, file)
		assert.False(file.closed)
		assert.DirExists(file.tempPath)
		assert.Empty(file.finalPath)
		assert.Empty(file.chunckPaths)
	})

	defer func(i item, t *testing.T) {
		if err := os.RemoveAll(i.tempPath); err != nil {
			t.Fatal("Fail to clean up directory", err)
		}
	}(file, t)

	t.Run("Test add one chunk", func(t *testing.T) {
		chunk := []byte("this is a chunk")
		err := sut.addChunk(file.id, 1, bytes.NewReader(chunk))

		i, ok := sut.files[file.id]
		assert := assert.New(t)
		assert.True(ok)
		assert.NoError(err)
		assert.False(i.closed)
		assert.Equal(1, len(i.chunckPaths))
		assert.FileExists(i.chunckPaths[0])
	})

	t.Run("Test another chunk", func(t *testing.T) {
		chunk := []byte(" and this is another chunk")
		err := sut.addChunk(file.id, 2, bytes.NewReader(chunk))

		i, ok := sut.files[file.id]
		assert := assert.New(t)
		assert.True(ok)
		assert.NoError(err)
		assert.False(i.closed)
		assert.Equal(2, len(i.chunckPaths))
		assert.FileExists(i.chunckPaths[0])
	})

	t.Run("Test finalize", func(t *testing.T) {
		err := sut.finalize(file.id)

		i, ok := sut.files[file.id]
		assert := assert.New(t)
		assert.True(ok)
		assert.NoError(err)
		assert.True(i.closed)
		assert.FileExists(i.finalPath)
	})

	t.Run("Test Download", func(t *testing.T) {
		var buf bytes.Buffer
		err := sut.read(file.id, &buf)
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal("this is a chunk and this is another chunk", buf.String())
	})
}
