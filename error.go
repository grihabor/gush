package compose

import (
	"fmt"
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func isError(t reflect.Type) bool {
	return t.Implements(errorInterface)
}

func canChainWithError(fn1 reflect.Type, fn2 reflect.Type) error {
	lastArg := fn1.Out(fn1.NumOut() - 1)
	if !isError(lastArg) {
		return fmt.Errorf("first function must return error as the last argument")
	}

	if fn1.NumOut()-1 != fn2.NumIn() {
		return fmt.Errorf(
			"first function returns %d arguments and error but second function takes %d arguments",
			fn1.NumOut()-1, fn2.NumIn(),
		)
	}

	for i := 0; i < fn2.NumIn(); i++ {
		outKind := fn1.Out(i).Kind()
		inKind := fn2.In(i).Kind()
		if outKind != inKind {
			return fmt.Errorf("arg #%d: %s != %s", i, outKind, inKind)
		}
	}
	return nil
}

func CanChainWithError(steps ...interface{}) error {
	fn := types(steps)
	for i := 0; i < len(steps)-1; i++ {
		idx1, idx2 := i, i+1
		err := canChainWithError(fn[idx1], fn[idx2])
		if err != nil {
			return fmt.Errorf(
				"failed to chain with error %v at index %d and %v at index %d: %w",
				fn[idx1], idx1, fn[idx2], idx2, err,
			)
		}
	}
	return nil
}

type LastArgError struct{}

func (r LastArgError) Stack(functions ...interface{}) (interface{}, error) {
	return SafeStack(functions...)
}

func (r LastArgError) Chain(functions ...interface{}) (interface{}, error) {
	return SafeStackWithError(functions...)
}

func ChainWithError(steps ...interface{}) interface{} {
	fn, err := SafeChainWithError(steps...)
	if err != nil {
		panic(err.Error())
	}
	return fn
}

func SafeChainWithError(steps ...interface{}) (interface{}, error) {
	if len(steps) < 2 {
		return nil, fmt.Errorf("chain with error can only work with 2 functions or more, got %v", steps)
	}
	err := CanChainWithError(steps...)
	if err != nil {
		return nil, fmt.Errorf("given functions can't be chained with error: %w", err)
	}

	first := reflect.TypeOf(steps[0])
	last := reflect.TypeOf(steps[len(steps)-1])
	lastLastArg := last.Out(last.NumOut() - 1)
	if !isError(lastLastArg) {
		return nil, fmt.Errorf("last function must return an error as it's last argument, got %v", last)
	}

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

	// precompute empty result for the case when err != nil
	emptyResult := make([]reflect.Value, 0, len(calls))
	for i := 0; i < last.NumOut()-1; i++ {
		emptyResult = append(emptyResult, reflect.New(last.Out(i)).Elem())
	}

	// build the resulting function
	return reflect.MakeFunc(resultFuncType, func(args []reflect.Value) []reflect.Value {
		var err reflect.Value
		for _, call := range calls {
			args = call(args)
			err = args[len(args)-1]
			if err.Interface() != nil {
				result := make([]reflect.Value, len(emptyResult))
				copy(result, emptyResult)
				return append(emptyResult, err)
			}
			args = args[:len(args)-1]
		}
		return append(args, err)
	}).Interface(), nil
}
