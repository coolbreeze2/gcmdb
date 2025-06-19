package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"diff", "-f", "../example/files"},
		{"diff", "-f", "../example/files/secret.yaml"},
		{"diff", "-f", "../example/files/datacenter.yaml"},
		{"diff", "-f", "../example/files/zone.yaml"},
		{"diff", "-f", "../example/files/namespace.yaml"},
		{"diff", "-f", "../example/files/scm.yaml"},
		{"diff", "-f", "../example/files/hostnode.yaml"},
		{"diff", "-f", "../example/files/helm_repository.yaml"},
		{"diff", "-f", "../example/files/container_registry.yaml"},
		{"diff", "-f", "../example/files/config_center.yaml"},
		{"diff", "-f", "../example/files/deploy_platform.yaml"},
		{"diff", "-f", "../example/files/project.yaml"},
		{"diff", "-f", "../example/files/app.yaml"},
		{"diff", "-f", "../example/files/deploy_template.yaml"},
		{"diff", "-f", "../example/files/resource_range.yaml"},
		{"diff", "-f", "../example/files/orchestration.yaml"},
		{"diff", "-f", "../example/files/appdeployment.yaml"},
	}

	ts := testServer()
	defer ts.Close()

	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
