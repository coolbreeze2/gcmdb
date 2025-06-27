package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRenderAppDeploymentFuncInvalid(t *testing.T) {
	route := chi.NewRouter()
	InstallApi(route, nil)

	rr := httptest.NewRecorder()
	path := fmt.Sprintf("%s/appdeployments/test/go-app/render", PathPrefix)
	req, _ := http.NewRequest("POST", path, nil)
	renderAppDeploymentFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)

	paramsByts, err := json.Marshal(map[string]any{"params": map[string]string{}})
	assert.NoError(t, err)
	req, _ = http.NewRequest("POST", path, bytes.NewBuffer(paramsByts))
	req.Header = http.Header{"Content-Type": {"application/json"}}
	renderAppDeploymentFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)
}

func TestRenderDeployTemplateFuncInvalid(t *testing.T) {
	route := chi.NewRouter()
	InstallApi(route, nil)

	rr := httptest.NewRecorder()
	path := fmt.Sprintf("%s/appdeployments/test/go-app/deploytemplate/render", PathPrefix)
	req, _ := http.NewRequest("POST", path, nil)
	renderDeployTemplateFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)

	req, _ = http.NewRequest("POST", path, bytes.NewBuffer([]byte(`{"params":{}}`)))
	req.Header = http.Header{"Content-Type": {"application/json"}}
	renderDeployTemplateFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)
}

func TestRunAppDeploymentFuncInvalid(t *testing.T) {
	route := chi.NewRouter()
	InstallApi(route, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", "go-app")
	rctx.URLParams.Add("namespace", "test")
	rctx.URLParams.Add("action", "release")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	runAppDeploymentFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)

	req, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"params":{}}`)))
	req.Header = http.Header{"Content-Type": {"application/json"}}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	runAppDeploymentFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)

	runAppDeploymentFunc()(rr, req)
	assert.Equal(t, rr.Code, 400)
}
