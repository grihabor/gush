package compose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanChain_True(t *testing.T) {
	err := CanChain(
		func() (int, float64) { return 0, 0 },
		func(int, float64) {},
	)
	assert.NoError(t, err)
}

func TestCanChain_False(t *testing.T) {
	err := CanChain(
		func() {},
		func(int, float64) {},
	)
	assert.Error(t, err)
}

func TestCanChain_Multiple(t *testing.T) {
	err := CanChain(
		func() (int, float64) { return 0, 0 },
		func(int, float64) {},
		func() {},
	)
	assert.NoError(t, err)
}

func TestChain(t *testing.T) {
	fn, err := SafeChain(
		func(a, b int) (float64, float64) { return float64(a), float64(b) },
		func(a, b float64) float64 { return a / b },
		func(c float64) int { return int(c) },
	)
	if !assert.NoError(t, err) {
		return
	}
	result := fn.(func(int, int) int)(6, 3)
	assert.Equal(t, 2, result)
}
