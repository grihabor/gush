package compose

import (
	"fmt"
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func isError(t reflect.Type) bool {
	return t.Implements(errorInterface)
}

func canChainWithError(args Args, fn1 reflect.Type, fn2 reflect.Type) error {
	fn1OutIndices := args.OutIndices(fn1)
	if len(fn1OutIndices) != fn2.NumIn() {
		return fmt.Errorf(
			"first function propagates %d of %d output args but second function takes %d args",
			len(fn1OutIndices), fn1.NumOut(), fn2.NumIn(),
		)
	}
	for i := 0; i < fn2.NumIn(); i++ {
		outKind := fn1.Out(fn1OutIndices[i]).Kind()
		inKind := fn2.In(i).Kind()
		if outKind != inKind {
			return fmt.Errorf("argument mismatch at index %d: %s != %s", i, outKind, inKind)
		}
	}
	return nil
}

func CanChainWithError(args Args, functions ...interface{}) error {
	fnTypes := types(functions)
	for _, fnType := range fnTypes {
		if err := args.CheckSpecialArgs(fnType); err != nil {
			return fmt.Errorf("failed to check special args: %w", err)
		}
	}
	for i := 0; i < len(functions)-1; i++ {
		idx1, idx2 := i, i+1
		err := canChainWithError(args, fnTypes[idx1], fnTypes[idx2])
		if err != nil {
			return fmt.Errorf(
				"failed to chain with error %v at index %d and %v at index %d: %w",
				fnTypes[idx1], idx1, fnTypes[idx2], idx2, err,
			)
		}
	}
	return nil
}

type Call func(values []reflect.Value) ([]reflect.Value, bool)

func ChainWithError(args Args, steps ...interface{}) interface{} {
	fn, err := SafeChainWithError(args, steps...)
	if err != nil {
		panic(err.Error())
	}
	return fn
}

type Args interface {
	CheckSpecialArgs(fn reflect.Type) error
	OutIndices(fn reflect.Type) []int
	BuildCall(fn interface{}) Call
	DataOk(data interface{}) bool
}

func SafeChainWithError(args Args, functions ...interface{}) (interface{}, error) {
	err := CanChainWithError(args, functions...)
	if err != nil {
		return nil, fmt.Errorf("given functions can't be chained with error: %w", err)
	}

	first := reflect.TypeOf(functions[0])
	last := reflect.TypeOf(functions[len(functions)-1])

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
	calls := make([]func(values []reflect.Value) ([]reflect.Value), len(functions))
	postprocessCalls := make([]Call, len(functions))
	for _, fn := range functions {
		calls = append(calls, func(values []reflect.Value) ([]reflect.Value) {
			return  reflect.ValueOf(fn).Call(values)
		})
		postprocessCalls = append(postprocessCalls, args.BuildCall(fn))
	}

	// precompute empty result for the case when err != nil
	emptyResult := make([]reflect.Value, 0, len(postprocessCalls))
	for i := 0; i < last.NumOut()-1; i++ {
		emptyResult = append(emptyResult, reflect.New(last.Out(i)).Elem())
	}

	// build the resulting function
	return reflect.MakeFunc(resultFuncType, func(values []reflect.Value) []reflect.Value {
		nextInputs := values
		var outputs []reflect.Value
		var ok bool
		for i, postprocess := range postprocessCalls {
			outputs = calls[i](nextInputs)
			nextInputs, ok = postprocess(values)
			if ok {
				continue
			}
			result := make([]reflect.Value, len(emptyResult))
			copy(result, emptyResult)
			return nextInputs
		}
		return outputs
	}).Interface(), nil
}
