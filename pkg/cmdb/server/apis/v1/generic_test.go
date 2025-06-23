package v1

import (
	"fmt"
	"gcmdb/pkg/cmdb/client"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testServer() *httptest.Server {
	r := NewRouter(nil)
	ts := httptest.NewServer(r)
	apiUrl := ts.URL + PathPrefix
	client.DefaultCMDBClient.ApiUrl = apiUrl
	return httptest.NewServer(r)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestCreate(t *testing.T) {
	// TODO:
}

func TestGetList(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	path := fmt.Sprintf("%s/%ss", PathPrefix, "secret")
	resp, _ := testRequest(t, ts, "GET", path, nil)
	assert.Equal(t, 200, resp.StatusCode, resp.Request.URL.String())
}

func TestGet(t *testing.T) {
	// TODO:
}

func TestUpdate(t *testing.T) {
	// TODO:
}

func TestDelete(t *testing.T) {
	// TODO:
}
