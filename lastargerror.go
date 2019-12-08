package compose

import (
	"fmt"
	"reflect"
)

type LastArgError struct{}

func (r LastArgError) BuildCall(fn interface{}) Call {
	return func(args []reflect.Value) ([]reflect.Value, []reflect.Value, bool) {
		err := args[len(args)-1]
		errInt := err.Interface()
		if errInt != nil {
			return args, errInt
		}
		return args[:len(args)-1], nil
	}
}

func (r LastArgError) OutIndices(fn reflect.Type) []int {
	n := fn.NumOut() - 1
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}
	return indices
}

func (r LastArgError) CheckSpecialArgs(fn reflect.Type) error {
	if fn.NumOut() < 1 {
		return fmt.Errorf("last returned argument must be an error, got %v", fn)
	}
	lastArg := fn.Out(fn.NumOut() - 1)
	if !isError(lastArg) {
		return fmt.Errorf("last returned argument must be an error, got %v", fn)
	}
	return nil
}

