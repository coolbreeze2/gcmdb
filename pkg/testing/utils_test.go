package testing

import (
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/utils"
	"reflect"
	"testing"
)

func TestMapToStruct(t *testing.T) {
	project := cmdb.NewProject()
	project.Metadata.Name = "Devops"
	project.Spec.NameInChain = "Dev"
	newProject := cmdb.Project{}
	m, err := utils.StructToMap(project)
	if err != nil {
		panic(err)
	}
	err = utils.MapToStruct(m, &newProject, reflect.TypeOf(project))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", newProject)
}

func TestArrayMapToStruct(t *testing.T) {
	var projects []cmdb.Project
	var newProjects []cmdb.Project
	p := cmdb.NewProject()
	p.Metadata.Name = "Devops"
	p.Spec.NameInChain = "Dev"
	projects = append(projects, *p)

	m, err := utils.StructToMap(projects)
	if err != nil {
		panic(err)
	}
	err = utils.MapToStruct(m["items"], &newProjects, reflect.TypeOf(p))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", newProjects)
}

func TestFieldByTag(t *testing.T) {
	project := cmdb.Project{}
	project.Metadata.Name = "Devops"
	metadata, ok := utils.FieldByTag(project, "json", "metadata")
	if ok {
		fmt.Printf("metadata:%v", metadata)
	}
}
