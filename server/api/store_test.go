package api

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepare(t *testing.T) {
	sut := NewStore()

	item, err := sut.prepare()

	assert := assert.New(t)
	assert.NoError(err)
	assert.False(item.closed)
	assert.DirExists(item.tempPath)
	defer os.RemoveAll(item.tempPath)
	assert.Empty(item.finalPath)
	assert.Empty(item.chunckPaths)
}

func TestAddChunk(t *testing.T) {
	sut := NewStore()
	item, err := sut.prepare()
	assert := assert.New(t)
	assert.NoError(err)

	chunk := []byte("this is a chunk")

	err = sut.addChunk(item.id, bytes.NewReader(chunk))

	item, ok := sut.files[item.id]
	assert.True(ok)
	assert.NoError(err)
	assert.False(item.closed)
	assert.DirExists(item.tempPath)
	defer os.RemoveAll(item.tempPath)
	assert.FileExists(item.chunckPaths[0])
}
