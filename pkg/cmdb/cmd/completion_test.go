package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompleteFunc(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	RootCmd.SetArgs([]string{"apply", "-f", "../example/files"})
	err := RootCmd.Execute()
	assert.NoError(t, err)
	options, _ := CompleteFunc(getCmd.Commands()[0], []string{}, "")
	assert.Less(t, 0, len(options))
}

func TestCompleteName(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	RootCmd.SetArgs([]string{"apply", "-f", "../example/files"})
	err := RootCmd.Execute()
	assert.NoError(t, err)

	options := completeName("app", "")
	assert.Less(t, 0, len(options))
}
