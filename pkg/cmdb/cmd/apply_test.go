package cmd

import (
	"goTool/global"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"apply", "-f", "../example/files/secret.yaml"},
		{"apply", "-f", "../example/files/datacenter.yaml"},
		{"apply", "-f", "../example/files/zone.yaml"},
		{"apply", "-f", "../example/files/namespace.yaml"},
		{"apply", "-f", "../example/files/scm.yaml"},
		{"apply", "-f", "../example/files/hostnode.yaml"},
		{"apply", "-f", "../example/files/helm_repository.yaml"},
		{"apply", "-f", "../example/files/container_registry.yaml"},
		{"apply", "-f", "../example/files/config_center.yaml"},
		{"apply", "-f", "../example/files/deploy_platform.yaml"},
		{"apply", "-f", "../example/files/project.yaml"},
		{"apply", "-f", "../example/files/app.yaml"},
		{"apply", "-f", "../example/files/deploy_template.yaml"},
		{"apply", "-f", "../example/files/resource_range.yaml"},
		{"apply", "-f", "../example/files/orchestration.yaml"},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestApplyInvalidAPIUrl(t *testing.T) {
	oldURL := global.ClientSetting.CMDB_API_URL
	global.ClientSetting.CMDB_API_URL = "http://a-bad-site.dev.com:8080/api/v1"
	RootCmd.SetArgs([]string{"apply", "-f", "../example/files/secret.yaml"})
	assertOsExit(t, Execute, 1)
	global.ClientSetting.CMDB_API_URL = oldURL
}

func TestApplyInvalid(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "resource")
	assert.NoError(t, err)
	defer os.Remove(tempDir)
	f, err := os.CreateTemp(tempDir, "secret.yaml")
	assert.NoError(t, err)
	_, err = f.Write([]byte(`apiVersion: v1alpha
kind: Secret
metadata:
  name: test
  extraFiled: xxxx
data:
  privateKey: 'MTIzNAo='`))
	assert.NoError(t, err)

	RootCmd.SetArgs([]string{"apply", "-f", tempDir})
	assertOsExit(t, Execute, 1)
}

func TestApplyUpdateResource(t *testing.T) {
	f, err := os.CreateTemp("", "secret.yaml")
	tempFilename := f.Name()
	defer os.Remove(tempFilename)
	assert.NoError(t, err)
	_, err = f.Write([]byte(`apiVersion: v1alpha
kind: Secret
metadata:
  name: test
data:
  privateKey: 'MTIzNAo='`))
	assert.NoError(t, err)

	cases := [][]string{
		{"apply", "-f", "../example/files/secret.yaml"},
		{"apply", "-f", tempFilename},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestApplyRefNotExist(t *testing.T) {
	f, err := os.CreateTemp("", "datacenter.yaml")
	tempFilename := f.Name()
	defer os.Remove(tempFilename)
	assert.NoError(t, err)
	_, err = f.Write([]byte(`apiVersion: v1alpha
kind: Datacenter
metadata:
  name: test
spec:
  provider: alibaba-cloud
  privateKey: a-not-exist-ref`))
	assert.NoError(t, err)

	RootCmd.SetArgs([]string{"apply", "-f", tempFilename})
	assertOsExit(t, Execute, 1)
}
