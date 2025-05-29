package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResource(t *testing.T) {

	resources := [][]string{
		{"secret", "test"},
		{"datacenter", "test"},
		{"zone", "test"},
		{"namespace", "test"},
		{"scm", "gitlab-test"},
		{"hostnode", "test"},
		{"helmrepository", "test"},
		{"containerregistry", "harbor-test"},
		{"configcenter", "apollo-test"},
		{"deployplatform", "test"},
		{"project", "go-devops"},
		{"app", "go-app"},
		{"deploytemplate", "docker-compose-test", "-n", "test"},
		{"resourcerange", "test", "-n", "test"},
		{"orchestration", "test"},
		{"appdeployment", "go-app", "-n", "test"},
	}

	cases := [][]string{
		{"apply", "-f", "../example/files"},
	}
	for _, r := range resources {
		ident := r[2:]
		c1 := append([]string{"get", r[0]}, ident...)
		cases = append(cases, c1)
		c2 := append([]string{"get", r[0], "-l", "x=y"}, ident...)
		cases = append(cases, c2)
		c3 := append([]string{"get", r[0], "-o", "yaml"}, ident...)
		cases = append(cases, c3)
		c4 := append([]string{"get", r[0], r[1], "-o", "yaml"}, ident...)
		cases = append(cases, c4)
		c5 := append([]string{"get", r[0], "-o", "json"}, ident...)
		cases = append(cases, c5)
		c6 := append([]string{"get", r[0], r[1], "-o", "json"}, ident...)
		cases = append(cases, c6)
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
		if flag := RootCmd.PersistentFlags().Lookup("namespace"); flag != nil {
			flag.Value.Set("")
		}
	}
}
