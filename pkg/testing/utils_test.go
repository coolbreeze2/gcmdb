package testing

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb/client"

	"testing"
)

func TestMapToStruct(t *testing.T) {
	project := client.Project{}
	project.Metadata.Name = "Devops"
	project.Spec.NameInChain = "Dev"
	m, err := json.Marshal(project)
	if err != nil {
		panic(err)
	}

	newProject := client.Project{}
	err = json.Unmarshal(m, &newProject)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", newProject)
}
