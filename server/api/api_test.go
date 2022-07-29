package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusEndpoint(t *testing.T) {
	api := NewApiServer("8888")

	ts := httptest.NewServer(api.handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/status", ts.URL), nil)
	if err != nil {
		t.Fatal("fail to create request", err)
	}

	client := http.Client{
		Timeout: time.Millisecond * 500,
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal("fail to make request", err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal("fail to read response", err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "ok", string(b))
}
