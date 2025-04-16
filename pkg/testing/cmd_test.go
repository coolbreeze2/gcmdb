package testing

import (
	"goTool/pkg/cmd"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetProject(t *testing.T) {
	args := []string{"get", "project"}
	cmd.RootCmd.SetArgs(args)
	err := cmd.RootCmd.Execute()
	assert.NoError(t, err)
}
