package compose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanChainWithError_True(t *testing.T) {
	err := CanChainWithError(
		LastArgError{},
		func() (int, float64, error) { return 0, 0, nil },
		func(int, float64) error { return nil },
	)
	assert.NoError(t, err)
}

func TestCanChainWithError_False(t *testing.T) {
	err := CanChainWithError(
		LastArgError{},
		func() (int, float64) { return 0, 0 },
		func(int, float64) {},
	)
	assert.Error(t, err)
}

func TestCanChainWithError_Multiple(t *testing.T) {
	err := CanChainWithError(
		LastArgError{},
		func() (int, float64, error) { return 0, 0, nil },
		func(int, float64) error { return nil },
		func() error { return nil },
	)
	assert.NoError(t, err)
}

func TestChainWithError(t *testing.T) {
	fn, err := SafeChainWithError(
		LastArgError{},
		func(a, b int) (float64, float64, error) { return float64(a), float64(b), nil },
		func(a, b float64) (float64, error) { return a / b, nil },
		func(c float64) (int, error) { return int(c), nil },
	)
	if !assert.NoError(t, err) {
		return
	}
	result, err := fn.(func(int, int) (int, error))(6, 3)
	assert.NoError(t, err)
	assert.Equal(t, 2, result)
}

func TestChainWithError_PropagateError(t *testing.T) {
	fn, err := SafeChainWithError(
		LastArgError{},
		func(a, b int) (float64, float64, error) { return float64(a), float64(b), fmt.Errorf("fail") },
		func(a, b float64) (float64, error) { return a / b, nil },
		func(c float64) (int, error) { return int(c), nil },
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = fn.(func(int, int) (int, error))(6, 3)
	assert.Error(t, err)
	assert.EqualError(t, err, "fail")
}
