package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlJoin(t *testing.T) {
	baseUrl := "http://123.com/api/v1"
	expectedUrl := "http://123.com/api/v1/apps/dev-app/"
	url := UrlJoin(baseUrl, "apps", "dev-app/")
	assert.Equal(t, expectedUrl, url)
}
