package compose

import (
	"fmt"
	"reflect"
)

func canChain(fn1 reflect.Type, fn2 reflect.Type) error {
	if fn1.NumOut() != fn2.NumIn() {
		return fmt.Errorf(
			"first function returns %d arguments but second function takes %d arguments",
			fn1.NumOut(), fn2.NumIn(),
		)
	}

	for i := 0; i < fn1.NumOut(); i++ {
		outKind := fn1.Out(i).Kind()
		inKind := fn2.In(i).Kind()
		if outKind != inKind {
			return fmt.Errorf("arg #%d: %s != %s", i, outKind, inKind)
		}
	}
	return nil
}

func CanChain(steps ...interface{}) error {
	fn := types(steps)
	if err := allFunctions(fn); err != nil {
		return fmt.Errorf("can't chain non functions: %w", err)
	}
	for i := 0; i < len(steps)-1; i++ {
		idx1, idx2 := i, i+1
		err := canChain(fn[idx1], fn[idx2])
		if err != nil {
			return fmt.Errorf(
				"failed to chain %v at index %d and %v at index %d: %w",
				fn[idx1], idx1, fn[idx2], idx2, err,
			)
		}
	}
	return nil
}

func Chain(steps ...interface{}) interface{} {
	fn, err := SafeChain(steps...)
	if err != nil {
		panic(err.Error())
	}
	return fn
}

type AllArgs struct{}

func (u AllArgs) Stack(functions ...interface{}) (interface{}, error) {
	return SafeStack(functions...)
}

func (u AllArgs) Chain(functions ...interface{}) (interface{}, error) {
	return SafeChain(functions...)
}

func SafeChain(steps ...interface{}) (interface{}, error) {
	err := CanChain(steps...)
	if err != nil {
		return nil, fmt.Errorf("given functions can't be chained: %w", err)
	}

	first := reflect.TypeOf(steps[0])
	last := reflect.TypeOf(steps[len(steps)-1])
	inFirst, err := in(first)
	if err != nil {
		return nil, fmt.Errorf("failed to get input types of the first function %v: %w", first, err)
	}
	outLast, err := out(last)
	if err != nil {
		return nil, fmt.Errorf("failed to get output types of the last function %v: %w", last, err)
	}
	resultFuncType := reflect.FuncOf(inFirst, outLast, false)

	// precompute calls to save time during execution
	calls := make([]func(in []reflect.Value) []reflect.Value, 0)
	for i := 0; i < len(steps); i++ {
		calls = append(calls, reflect.ValueOf(steps[i]).Call)
	}

	// build the resulting function
	return reflect.MakeFunc(resultFuncType, func(args []reflect.Value) []reflect.Value {
		for _, call := range calls {
			args = call(args)
		}
		return args
	}).Interface(), nil
}
