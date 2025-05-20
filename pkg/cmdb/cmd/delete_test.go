package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files/project.yaml"},
		{"apply", "-f", "../example/files/app.yaml"},
		{"delete", "app", "go-app"},
		{"delete", "project", "go-devops"},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
