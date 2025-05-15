package testing

import (
	"fmt"
	"goTool/pkg/cmdb/client"
	"testing"
)

func TestUrlJoin(t *testing.T) {
	baseUrl := "http://123.com/api/v1"
	url, err := client.UrlJoin(baseUrl, "apps", "dev-app/")
	if err != nil {
		panic(err)
	}
	fmt.Printf("url: %v", url)
}
