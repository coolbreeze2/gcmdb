package cmd

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	RootCmd.SetArgs([]string{})
	RootCmd.Execute()
}

func TestExcute(t *testing.T) {
	Execute()
}

func TestNotExistCmd(t *testing.T) {
	RootCmd.SetArgs([]string{"not-exsit-cmd"})
	assertOsExit(t, Execute, 1)
}

func assertOsExit(t *testing.T, f assert.PanicTestFunc, code int, msgAndArgs ...any) {
	msg := fmt.Sprintf("os.Exit(%s) called", strconv.Itoa(code))
	fakeExit := func(int) {
		panic(msg)
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()
	assert.PanicsWithValue(t, msg, f, msgAndArgs...)
}
