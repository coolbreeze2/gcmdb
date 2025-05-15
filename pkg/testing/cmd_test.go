package testing

import (
	"goTool/pkg/cmd"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProject(t *testing.T) {
	cases := [][]string{
		{"get", "project"},
		{"get", "project", "devops"},
		{"get", "project", "-o", "yaml"},
		{"get", "project", "devops", "-o", "yaml"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestApplyProject(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "example/files/project.yaml"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestGetApp(t *testing.T) {
	cases := [][]string{
		{"get", "app"},
		{"get", "app", "dev-app"},
		{"get", "app", "-o", "yaml"},
		{"get", "app", "dev-app", "-o", "yaml"},
	}
	for i := range cases {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}
