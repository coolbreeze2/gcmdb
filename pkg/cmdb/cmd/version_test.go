package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	RootCmd.SetArgs([]string{"version"})
	err := RootCmd.Execute()
	assert.NoError(t, err)
}
