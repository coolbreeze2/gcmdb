package cmd

import (
	"gcmdb/pkg/cmdb/client"
	apiv1 "gcmdb/pkg/cmdb/server/apis/v1"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testServer() *httptest.Server {
	r := apiv1.NewRouter(nil)
	ts := httptest.NewServer(r)
	apiUrl := ts.URL + apiv1.PathPrefix
	client.DefaultCMDBClient.ApiUrl = apiUrl
	return ts
}

func TestVersion(t *testing.T) {
	RootCmd.SetArgs([]string{"version"})
	err := RootCmd.Execute()
	assert.NoError(t, err)
}
