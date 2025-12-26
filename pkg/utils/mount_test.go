package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMounted(t *testing.T) {
	mounted, err := IsMounted("/")
	assert.NoError(t, err)
	assert.True(t, mounted)

	mounted, err = IsMounted("/bin")
	assert.NoError(t, err)
	assert.False(t, mounted)

	_, err = IsMounted("/non-exist")
	assert.Error(t, err)
}
