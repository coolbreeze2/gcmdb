package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsErrorCode(t *testing.T) {
	assert.Equal(t, isErrCode(nil, 1), false)
	assert.Equal(t, isErrCode(os.ErrExist, 1), false)
}
