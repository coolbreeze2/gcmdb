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
	}
	for i := 0; i < len(cases); i++ {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestGetApp(t *testing.T) {
	cases := [][]string{
		{"get", "app"},
		{"get", "app", "dev-app"},
	}
	for i := 0; i < len(cases); i++ {
		cmd.RootCmd.SetArgs(cases[i])
		err := cmd.RootCmd.Execute()
		assert.NoError(t, err)
	}
}
