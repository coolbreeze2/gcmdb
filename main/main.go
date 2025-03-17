package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"goTool/cmdb"
	"io"
	"log"
	"net/http"
)

var name string

type pingDataFormat struct {
	UserAccessToken          string `json:"userAccessToken"`
	UploadStartTimeInSeconds int    `json:"uploadStartTimeInSeconds"`
	UploadEndTimeInSeconds   int    `json:"uploadEndTimeInSeconds"`
	CallbackURL              string `json:"callbackURL"`
}

func httpGet(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	fmt.Printf("%s", body)
}

func testMarshal() {
	project := cmdb.NewProject()
	project.Metadata.Name = "goTool"
	project.Spec.NameInChain = "devops"

	jsonData, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}

	pUnmarshaled := &cmdb.Project{}
	if err = json.Unmarshal([]byte(jsonData), &pUnmarshaled); err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))

	fmt.Printf("Obj Project: %v", *pUnmarshaled)
}

func testList() {
	p := cmdb.NewProject()
	projects := p.List()
	jsonData, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) <= 0 {
		return
	}

	switch args[0] {
	case "go":
		goCmd := flag.NewFlagSet("go", flag.ExitOnError)
		goCmd.StringVar(&name, "name", "Go 语言", "帮助信息")
		_ = goCmd.Parse(args[1:])
		testList()
	case "php":
		phpCmd := flag.NewFlagSet("php", flag.ExitOnError)
		phpCmd.StringVar(&name, "n", "PHP 语言", "帮助信息")
		_ = phpCmd.Parse(args[1:])
		url := "https://www.baidu.com"
		httpGet(url)
	}

	log.Printf("name: %s", name)
}
