package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyResource(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files/project.yaml"},
		{"apply", "-f", "../example/files/app.yaml"},
	}
	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}
