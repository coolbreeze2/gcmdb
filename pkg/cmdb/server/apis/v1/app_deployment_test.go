package v1

import (
	"fmt"
	"gcmdb/pkg/cmdb/cmd"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func testInit(t *testing.T) {
	cmd.RootCmd.SetArgs([]string{"apply", "-f", "../../../example/files"})
	err := cmd.RootCmd.Execute()
	assert.NoError(t, err)
}

func TestResolveAppDeployment(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	testInit(t)

	appDeploy, err := resolveAppDeployment("go-app", "test", map[string]any{})
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	out, err := yaml.MarshalWithOptions(appDeploy, yaml.AutoInt())
	fmt.Println(string(out))
}

func TestResolveDeployTemplate(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	testInit(t)

	appDeploy, err := resolveDeployTemplate("go-app", "test", map[string]any{})
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	out, err := yaml.MarshalWithOptions(appDeploy, yaml.AutoInt(), yaml.UseLiteralStyleIfMultiline(true))
	fmt.Println(string(out))
}
