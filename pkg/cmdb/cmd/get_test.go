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
		{"project", "go-devops"},
		{"app", "go-app"},
	}

	cases := [][]string{
		{"apply", "-f", "../example/files"},
	}
	for _, r := range resources {
		cases = append(cases, []string{"get", r[0]})
		cases = append(cases, []string{"get", r[0], "-l", "x=y"})
		cases = append(cases, []string{"get", r[0], "-o", "yaml"})
		cases = append(cases, []string{"get", r[0], r[1], "-o", "yaml"})
		cases = append(cases, []string{"get", r[0], "-o", "json"})
		cases = append(cases, []string{"get", r[0], r[1], "-o", "json"})
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
