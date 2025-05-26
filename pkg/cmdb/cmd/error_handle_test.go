package cmd

import (
	"goTool/pkg/cmdb"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestCheckError(t *testing.T) {
	fakeExit := func(code int) {
		assert.Equal(t, 1, code)
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	errs := []error{
		cmdb.ResourceNotFoundError{},
		cmdb.ResourceValidateError{},
		cmdb.ResourceAlreadyExistError{},
		cmdb.ResourceReferencedError{},
		cmdb.ServerError{},
	}
	for _, err := range errs {
		CheckError(err)
	}
}
