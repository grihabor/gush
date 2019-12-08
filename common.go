package compose

import (
	"fmt"
	"reflect"
)

func types(steps []interface{}) []reflect.Type {
	fn := make([]reflect.Type, 0, len(steps))
	for _, step := range steps {
		fn = append(fn, reflect.TypeOf(step))
	}
	return fn
}

func in(fn reflect.Type) ([]reflect.Type, error) {
	if fn.Kind() != reflect.Func {
		return nil, fmt.Errorf("given type is not a function: %v", fn)
	}
	numIn := fn.NumIn()
	result := make([]reflect.Type, 0, numIn)
	for i := 0; i < numIn; i++ {
		result = append(result, fn.In(i))
	}
	return result, nil
}

func out(fn reflect.Type) ([]reflect.Type, error) {
	if fn.Kind() != reflect.Func {
		return nil, fmt.Errorf("given type is not a function: %v", fn)
	}
	numOut := fn.NumOut()
	result := make([]reflect.Type, 0, numOut)
	for i := 0; i < numOut; i++ {
		result = append(result, fn.Out(i))
	}
	return result, nil
}
