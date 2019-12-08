package builder

import (
	"fmt"
)

type GraphBuilder struct {
	nodes []*Node
}

func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{}
}

func (g *GraphBuilder) Node(fn interface{}) *Node {
	node := &Node{fn: fn}
	g.nodes = append(g.nodes, node)
	return node
}

func (g *GraphBuilder) SafeBuild() (*Graph, error) {
	p := &Graph{}
	for nodeIndex, node := range g.nodes {
		nodeFn := node.fn
		i, err := p.insert(nodeFn)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to insert node %v at #%d: %w",
				nodeFn, nodeIndex, err,
			)
		}
		for inputIndex, input := range node.inputs {
			j, err := p.insert(input)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to insert input %v at #%d for node %v at #%d: %w",
					input, inputIndex, nodeFn, nodeIndex, err,
				)
			}
			// p.edge[i] is created during p.insert
			p.edge[i] = append(p.edge[i], j)
		}
	}
	return p, nil
}

func (g *GraphBuilder) Build() (*Graph, error) {
	graph, err := g.SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("failed to SafeBuild the graph: %w", err)
	}
	return graph, nil
}

func (f *Node) Inputs(inputs ...interface{}) {
	f.inputs = inputs
}

type Node struct {
	fn     interface{}
	inputs []interface{}
}
