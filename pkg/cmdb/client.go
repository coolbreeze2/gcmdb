package cmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func (p *Resource) List() []Resource {
	api_url := os.Getenv("CMDB_API_URL")
	url := api_url + "/" + strings.ToLower(p.Kind) + "s/"
	body := httpGet(url)
	resources := &[]Resource{}
	if err := json.Unmarshal([]byte(body), &resources); err != nil {
		panic(err)
	}
	return *resources
}

func httpGet(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	fmt.Printf("%s", body)
	return body
}
