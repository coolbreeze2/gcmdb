package cmdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceWithKindBad(t *testing.T) {
	_, err := NewResourceWithKind("bad-kind")
	assert.EqualError(t, err, ResourceTypeError{Kind: "bad-kind"}.Error())
}
