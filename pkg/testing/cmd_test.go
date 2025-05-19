package testing

import (
	"goTool/pkg/cmd"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "example/files/project.yaml"},
		{"apply", "-f", "example/files/app.yaml"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestGetResource(t *testing.T) {
	cases := [][]string{
		{"get", "project"},
		{"get", "project", "go-devops"},
		{"get", "project", "-o", "yaml"},
		{"get", "project", "go-devops", "-o", "yaml"},
		{"get", "app"},
		{"get", "app", "go-app"},
		{"get", "app", "-o", "yaml"},
		{"get", "app", "go-app", "-o", "yaml"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestDeleteResource(t *testing.T) {
	cases := [][]string{
		{"delete", "app", "go-app"},
		{"delete", "project", "go-devops"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}
