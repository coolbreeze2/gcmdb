package main

import (
	apiv1 "goTool/pkg/cmdb/server/apis/v1"
	"net/http"
)

func main() {
	r := apiv1.NewRouter(nil)
	http.ListenAndServe(":3333", r)
}
