package compose

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grihabor/gush/builder"
)

func TestNewGraph_DonorCalledOnce(t *testing.T) {
	gb := builder.NewGraphBuilder()

	var callCount int
	src := func() int {
		callCount += 1
		return 42
	}
	gb.Node(func(a int) int { return a / 2 }).Inputs(src)
	gb.Node(func(a int) int { return a / 3 }).Inputs(src)

	g, err := gb.Build()
	if !assert.NoError(t, err) {
		return
	}
	fn, err := SafeCompile(g, AllArgs{})
	if !assert.NoError(t, err) {
		return
	}
	a, b := fn.(func() (int, int))()
	assert.Equal(t, 21, a)
	assert.Equal(t, 14, b)
	assert.Equal(t, 1, callCount)
}

func TestNewGraph_Chain(t *testing.T) {
	gb := builder.NewGraphBuilder()

	one := func(numbers []int) []int {
		return append(numbers, 1)
	}
	two := func(numbers []int) []int {
		return append(numbers, 2)
	}
	three := func(numbers []int) []int {
		return append(numbers, 3)
	}
	gb.Node(two).Inputs(one)
	gb.Node(three).Inputs(two)

	g, err := gb.Build()
	if !assert.NoError(t, err) {
		return
	}
	fn, err := SafeCompile(g, AllArgs{})
	if !assert.NoError(t, err) {
		return
	}
	fun := fn.(func([]int) []int)
	result := fun(make([]int, 0))
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestNewGraph_MultipleOutputArgs(t *testing.T) {
	gb := builder.NewGraphBuilder()

	ab := func() (int, float64) { return 13, 5.5 }
	c := func() float32 { return 7.5 }
	gb.Node(func(a int, b float64, c float32) float64 {
		return float64(a) + b + float64(c)
	}).Inputs(ab, c)
	gb.Node(func(c float32, a int, b float64) float64 {
		return float64(a) + b + float64(c)
	}).Inputs(c, ab)

	g, err := gb.Build()
	if !assert.NoError(t, err) {
		return
	}
	fn, err := SafeCompile(g, AllArgs{})
	if !assert.NoError(t, err) {
		return
	}
	fun := fn.(func() (float64, float64))
	x, y := fun()
	assert.Equal(t, float64(26), x)
	assert.Equal(t, float64(26), y)
}
