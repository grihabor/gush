package compose

import (
	// "fmt"
	"fmt"
	"reflect"
)

// readyToBeCalculated returns all the nodes which already has all inputs calculated
func readyToBeCalculated(g G, calculated []int) []int {

	isCalculated := func(i int) bool {
		for _, calculatedIndex := range calculated {
			if i == calculatedIndex {
				return true
			}
		}
		return false
	}

	allInputsCalculated := func(inputs []int) bool {
		for _, input := range inputs {
			if !isCalculated(input) {
				return false
			}
		}
		return true
	}

	result := make([]int, 0)
	g.ForEachNode(func(i int, inputs []int) {
		if isCalculated(i) {
			return
		}
		if allInputsCalculated(inputs) {
			result = append(result, i)
		}
	})
	return result
}

type G interface {
	NodeCount() int
	Nodes(indices []int) []interface{}
	Inputs(int) []int
	ForEachNode(func(int, []int))
}

func glue(g G, donorLayerIndices []int, recipientLayerIndices []int) (interface{}, error) {
	donorLayer, recipientLayer := g.Nodes(donorLayerIndices), g.Nodes(recipientLayerIndices)
	donorLayerTypes, recipientLayerTypes := types(donorLayer), types(recipientLayer)
	donorLayerOutputTypes, err := mapEach(out, donorLayerTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve donor layer %v output types: %w", donorLayerTypes, err)
	}
	recipientLayerInputTypes, err := mapEach(in, recipientLayerTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve recipient layer %v input types: %w", donorLayerTypes, err)
	}
	layer1Flatten := flatten(donorLayerOutputTypes)
	layer2Flatten := flatten(recipientLayerInputTypes)

	offsets := make([]int, 1)
	indicesMapping := make(map[int]int)
	for offsetIndex, types := range donorLayerOutputTypes {
		next := offsets[len(offsets)-1] + len(types)
		offsets = append(offsets, next)
		indicesMapping[donorLayerIndices[offsetIndex]] = offsetIndex
	}

	resultFuncType := reflect.FuncOf(layer1Flatten, layer2Flatten, false)
	return reflect.MakeFunc(resultFuncType, func(args []reflect.Value) []reflect.Value {
		result := make([]reflect.Value, 0, len(recipientLayerInputTypes))
		for _, idx := range recipientLayerIndices {
			inputIndices := g.Inputs(idx)
			for _, inputIndex := range inputIndices {
				offsetIndex := indicesMapping[inputIndex]
				result = append(result, args[offsets[offsetIndex]:offsets[offsetIndex+1]]...)
			}
		}
		return result
	}).Interface(), nil
}

func node(g G, idx int) interface{} {
	return g.Nodes([]int{idx})[0]
}

type Ops interface {
	Stack(...interface{}) (interface{}, error)
	Chain(...interface{}) (interface{}, error)
}

// build resulting function
func SafeCompile(g G, ops Ops) (interface{}, error) {
	calculated := make([]int, 0)
	indicesToBeChained := make([][]int, 0)
	for len(calculated) < g.NodeCount() {
		readyIndices := readyToBeCalculated(g, calculated)
		indicesToBeChained = append(indicesToBeChained, readyIndices)
		calculated = append(calculated, readyIndices...)
	}
	toBeChained := make([]interface{}, 0)
	for i, indices := range indicesToBeChained {
		if i > 0 {
			prevIndices := indicesToBeChained[i-1]
			glued, err := glue(g, prevIndices, indices)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to glue function %v at %d and function %v at %d",
					node(g, i-1), i-1, node(g, i), i,
				)
			}
			toBeChained = append(toBeChained, glued)
		}
		ready := g.Nodes(indices)
		stacked, err := ops.Stack(ready...)
		if err != nil {
			return nil, fmt.Errorf("failed to stack functions %v: %w", types(ready), err)
		}
		toBeChained = append(toBeChained, stacked)
	}
	chained, err := ops.Chain(toBeChained...)
	if err != nil {
		return nil, fmt.Errorf("failed to chain %v: %w", types(toBeChained), err)
	}
	return chained, nil
}

func Compile(g G, ops Ops) interface{} {
	result, err := SafeCompile(g, ops)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return result
}
