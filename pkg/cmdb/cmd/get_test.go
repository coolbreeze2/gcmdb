package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"get", "scm"},
		{"get", "scm", "-o", "yaml"},
		{"get", "scm", "-o", "json"},
		{"get", "scm", "gitlab-test"},
		{"get", "datacenter"},
		{"get", "datacenter", "-o", "yaml"},
		{"get", "datacenter", "-o", "json"},
		{"get", "datacenter", "test"},
		{"get", "secret"},
		{"get", "secret", "-o", "yaml"},
		{"get", "secret", "-o", "json"},
		{"get", "secret", "test"},
		{"get", "project"},
		{"get", "project", "go-devops"},
		{"get", "project", "-o", "yaml"},
		{"get", "project", "-o", "json"},
		{"get", "project", "go-devops", "-o", "yaml"},
		{"get", "app"},
		{"get", "app", "go-app"},
		{"get", "app", "-o", "yaml"},
		{"get", "app", "-o", "json"},
		{"get", "app", "go-app", "-o", "yaml"},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
