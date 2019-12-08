package compose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	fn, err := SafeStack(
		func(a int) int { return a + 1 },
		func(a int) int { return a * 2 },
	)
	if !assert.NoError(t, err) {
		return
	}
	a, b := fn.(func(int, int) (int, int))(5, 5)
	assert.Equal(t, 6, a)
	assert.Equal(t, 10, b)
}
