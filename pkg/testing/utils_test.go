package testing

import (
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/client"
	"goTool/pkg/utils"
	"reflect"
	"testing"
)

func TestMapToStruct(t *testing.T) {
	project := client.Project{}
	project.Metadata.Name = "Devops"
	project.Spec.NameInChain = "Dev"
	newProject := client.NewProject()
	m, err := utils.StructToMap(project)
	if err != nil {
		panic(err)
	}
	err = utils.MapToStruct(m, newProject, reflect.TypeOf(project))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", newProject)
}

// func TestArrayMapToStruct(t *testing.T) {
// 	var projects []cmdb.Project
// 	var newProjects []cmdb.Project
// 	p := cmdb.Project{}
// 	p.Metadata.Name = "Devops"
// 	p.Spec.NameInChain = "Dev"
// 	projects = append(projects, p)

// 	m, err := utils.StructToMap(projects)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = utils.MapToStruct(m["items"], newProjects, reflect.TypeOf(p))
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%v", newProjects)
// }

func TestFieldByTag(t *testing.T) {
	project := cmdb.Project{}
	project.Metadata.Name = "Devops"
	metadata, ok := utils.FieldByTag(project, "json", "metadata")
	if ok {
		fmt.Printf("metadata:%v", metadata)
	}
}
