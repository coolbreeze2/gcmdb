package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteResource(t *testing.T) {
	// 倒序删除
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"delete", "appdeployment", "go-app", "-n", "test"},
		{"delete", "orchestration", "test"},
		{"delete", "resourcerange", "test", "-n", "test"},
		{"delete", "deploytemplate", "docker-compose-test", "-n", "test"},
		{"delete", "app", "go-app"},
		{"delete", "project", "go-devops"},
		{"delete", "deployplatform", "test"},
		{"delete", "configcenter", "apollo-test"},
		{"delete", "containerregistry", "harbor-test"},
		{"delete", "helmrepository", "test"},
		{"delete", "hostnode", "test"},
		{"delete", "scm", "gitlab-test"},
		{"delete", "namespace", "test"},
		{"delete", "zone", "test"},
		{"delete", "datacenter", "test"},
		{"delete", "secret", "test"},
	}

	ts := testServer()
	defer ts.Close()

	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
		if flag := RootCmd.PersistentFlags().Lookup("namespace"); flag != nil {
			flag.Value.Set("")
		}
	}
}
