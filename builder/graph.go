package builder

import (
	"fmt"
	"reflect"
)

type Graph struct {
	// node store functions of the graph
	node []interface{}
	// edge describes function inputs in the graph:
	// inputs for node[i] which takes n inputs: edge[i][0], ..., edge[i][n]
	edge [][]int
	// function to use to chain functions
	chain func(steps ...interface{}) (interface{}, error)
	// function to use to stack functions
	stack func(steps ...interface{}) (interface{}, error)
}

func (g *Graph) NodeCount() int {
	return len(g.node)
}

func (g *Graph) Inputs(idx int) []int {
	return g.edge[idx]
}

func (g *Graph) ForEachNode(callback func(int, []int)) {
	for i, inputs := range g.edge {
		callback(i, inputs)
	}
}

// get corresponding nodes for given indices
func (g *Graph) Nodes(indices []int) []interface{} {
	nodes := make([]interface{}, 0, len(indices))
	for _, idx := range indices {
		nodes = append(nodes, g.node[idx])
	}
	return nodes
}

// insert returns an index of the inserted function
func (g *Graph) insert(fn interface{}) (int, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return 0, fmt.Errorf("not a function: %v", fnType)
	}
	// search for the fn in the list nodes
	for i, nodeFn := range g.node {
		if reflect.ValueOf(fn) == reflect.ValueOf(nodeFn) {
			return i, nil
		}
	}
	// insert if we failed to find it
	g.node = append(g.node, fn)
	// align edge array so that indices match
	g.edge = append(g.edge, make([]int, 0))
	return len(g.node) - 1, nil
}
