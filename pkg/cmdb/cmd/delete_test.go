package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteResource(t *testing.T) {
	// 倒序删除
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"delete", "app", "go-app"},
		{"delete", "project", "go-devops"},
		{"delete", "scm", "gitlab-test"},
		{"delete", "namespace", "test"},
		{"delete", "zone", "test"},
		{"delete", "datacenter", "test"},
		{"delete", "secret", "test"},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
