package v1

import (
	"context"
	"gcmdb/global"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/server/storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testInvalidSetup() (context.Context, *storage.Store, *clientv3.Client) {
	client, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"invalid-endpoint-url"},
	})
	store := storage.New(client, global.StoragePathPrefix)
	ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	return ctx, store, client
}

func TestAddGenericApiInvalidKind(t *testing.T) {
	defer func() {
		assert.IsType(t, cmdb.ResourceTypeError{}, recover())
	}()
	r := chi.NewRouter()
	addGenericApi(r, "invalid-kind")
}

func TestHandleWithInvalidStore(t *testing.T) {
	ctx, store, _ := testInvalidSetup()
	route := chi.NewRouter()
	InstallApi(route, store)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)
	countFunc("invalid-kind")(rr, req)
	assert.Equal(t, rr.Code, 500)

	getNamesFunc("invalid-kind")(rr, req)
	assert.Equal(t, rr.Code, 500)

	getListFunc("invalid-kind")(rr, req)
	assert.Equal(t, rr.Code, 500)
}

func TestCreateFuncInvalid(t *testing.T) {
	route := chi.NewRouter()
	InstallApi(route, nil)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	createFunc("invalid-kind")(rr, req)
	assert.Equal(t, rr.Code, 400)

	createFunc("secret")(rr, req)
	assert.Equal(t, rr.Code, 400)
}

func TestUpdateFuncInvalid(t *testing.T) {
	route := chi.NewRouter()
	InstallApi(route, nil)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	updateFunc("invalid-kind")(rr, req)
	assert.Equal(t, rr.Code, 400)

	updateFunc("secret")(rr, req)
	assert.Equal(t, rr.Code, 400)
}
