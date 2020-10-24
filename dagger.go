package dagger

import (
	"github.com/autom8ter/dagger/primitive"
)

var globalGraph = primitive.NewGraph()

func EdgeTypes() []string {
	return globalGraph.EdgeTypes()
}

func NodeTypes() []string {
	return globalGraph.NodeTypes()
}

func GetNode(id primitive.TypedID) (*Node, bool) {
	n, ok := globalGraph.GetNode(id)
	if !ok {
		return nil, false
	}
	return &Node{n}, true
}

func RangeNodeTypes(typ primitive.Type, fn func(n *Node) bool) {
	globalGraph.RangeNodeTypes(typ, func(n primitive.Node) bool {
		return fn(&Node{n})
	})
}

func RangeNodes(fn func(n *Node) bool) {
	globalGraph.RangeNodes(func(n primitive.Node) bool {
		return fn(&Node{n})
	})
}

func RangeEdges(fn func(e *primitive.Edge) bool) {
	globalGraph.RangeEdges(fn)
}

func RangeEdgeTypes(edgeType primitive.Type, fn func(e *primitive.Edge) bool) {
	globalGraph.RangeEdgeTypes(edgeType, fn)
}

func HasNode(id primitive.TypedID) bool {
	return globalGraph.HasNode(id)
}

func Close() {
	globalGraph.Close()
}
