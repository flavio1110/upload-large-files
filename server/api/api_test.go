package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusEndpoint(t *testing.T) {
	api := NewApiServer("8888")
	ts := httptest.NewServer(api.handler)
	defer ts.Close()
	client := http.Client{
		Timeout: time.Millisecond * 500,
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/status", ts.URL), nil)
	b := DoRequest(t, client, req, http.StatusOK)
	assert.Equal(t, "ok", string(b))
}

func TestUploadFlow(t *testing.T) {
	api := NewApiServer("8888")
	ts := httptest.NewServer(api.handler)
	defer ts.Close()
	if err := os.Mkdir("temp", 0777); err != nil {
		t.Fatal("failed to create temp folder", err)
	}
	defer func() {
		if err := os.RemoveAll("temp"); err != nil {
			t.Fatal("failed to delete temp folder")
		}
	}()
	client := http.Client{
		Timeout: time.Millisecond * 500,
	}
	var fileId string
	t.Run("Test prepare", func(t *testing.T) {
		body := `{
			"name" : "text.txt",
			"content_type" : "text/plain"
		}`
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/file/prepare", ts.URL),
			bytes.NewBufferString(body))

		b := DoRequest(t, client, req, http.StatusOK)

		resp := struct {
			Id string `json:"id"`
		}{}

		if err := json.Unmarshal(b, &resp); err != nil {
			t.Fatal("parse response", err)
		}
		assert.NotEmpty(t, resp.Id)
		fileId = resp.Id
	})
	t.Run("Test add chunks", func(t *testing.T) {
		fileDir, _ := os.Getwd()
		filePath := path.Join(fileDir, "temp", "chunk1")

		file, _ := os.Create(filePath)
		defer file.Close()
		fmt.Fprint(file, "first part")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("chunk", filepath.Base(file.Name()))
		file.Seek(0, io.SeekStart)
		io.Copy(part, file)
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/file/add-chunk/%s/1", ts.URL, fileId),
			body)
		req.Header.Add("Content-Type", writer.FormDataContentType())

		DoRequest(t, client, req, http.StatusCreated)

		filePath = path.Join(fileDir, "/temp", "chunk2")

		file, _ = os.Create(filePath)
		defer file.Close()
		defer os.Remove(filePath)
		fmt.Fprint(file, ", second part")

		body = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
		part, _ = writer.CreateFormFile("chunk", filepath.Base(file.Name()))
		file.Seek(0, io.SeekStart)
		io.Copy(part, file)
		writer.Close()

		req, _ = http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/file/add-chunk/%s/2", ts.URL, fileId),
			body)
		req.Header.Add("Content-Type", writer.FormDataContentType())

		DoRequest(t, client, req, http.StatusCreated)
	})

	t.Run("Test finalize", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/file/finalize/%s", ts.URL, fileId), nil)

		DoRequest(t, client, req, http.StatusOK)
	})

	t.Run("Test download", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/file/download/%s", ts.URL, fileId), nil)

		b := DoRequest(t, client, req, http.StatusOK)

		assert.Equal(t, "first part, second part", string(b))
	})
}

func DoRequest(t *testing.T, client http.Client, req *http.Request, expectedCode int) []byte {
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("fail to make request", err)
	}

	b, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal("fail to read response", err)
	}

	assert.Equal(t, expectedCode, res.StatusCode)
	return b
}
