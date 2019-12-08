package compose

import (
	"fmt"
	"reflect"
)

func mapEach(
	mapFunction func(reflect.Type) ([]reflect.Type, error),
	functions []reflect.Type,
) ([][]reflect.Type, error) {
	result := make([][]reflect.Type, 0)
	for i, fn := range functions {
		item, err := mapFunction(fn)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to apply given mapFunction to item #%d: %w",
				i, err,
			)
		}
		result = append(result, item)
	}
	return result, nil
}

func flatten(
	argsList [][]reflect.Type,
) []reflect.Type {
	result := make([]reflect.Type, 0)
	for _, args := range argsList {
		result = append(result, args...)
	}
	return result
}

func Stack(steps ...interface{}) interface{} {
	result, err := SafeStack(steps...)
	if err != nil {
		panic(err.Error())
	}
	return result
}

func allFunctions(types []reflect.Type) error {
	for _, typ := range types {
		if typ.Kind() != reflect.Func {
			return fmt.Errorf("expected %v, got %v", reflect.Func, typ)
		}
	}
	return nil
}

func SafeStack(steps ...interface{}) (interface{}, error) {
	functions := types(steps)
	if err := allFunctions(functions); err != nil {
		return nil, fmt.Errorf("can't stack non functions %v: %w", steps, err)
	}
	inputTypes, err := mapEach(in, functions)
	if err != nil {
		return nil, fmt.Errorf("failed to get functions input types: %w", err)
	}
	outputTypes, err := mapEach(out, functions)
	if err != nil {
		return nil, fmt.Errorf("failed to get functions output types: %w", err)
	}
	result := reflect.FuncOf(flatten(inputTypes), flatten(outputTypes), false)

	// precompute calls to save time during execution
	calls := make([]func(in []reflect.Value) []reflect.Value, 0)
	for i := 0; i < len(steps); i++ {
		calls = append(calls, reflect.ValueOf(steps[i]).Call)
	}

	return reflect.MakeFunc(result, func(args []reflect.Value) (results []reflect.Value) {
		start := 0
		outputs := make([]reflect.Value, 0)
		for i, call := range calls {
			inputs := args[start : start+len(inputTypes[i])]
			outputs = append(outputs, call(inputs)...)
			start += len(inputTypes[i])
		}
		return outputs
	}).Interface(), nil
}

func SafeStackWithError(steps ...interface{}) (interface{}, error) {
	panic("not implemented")
}
