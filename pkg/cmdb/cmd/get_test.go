package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files/project.yaml"},
		{"apply", "-f", "../example/files/app.yaml"},
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
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
