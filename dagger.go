package dagger

import (
	"github.com/autom8ter/dagger/primitive"
	"sort"
)

var globalGraph = primitive.NewGraph()

// EdgeTypes returns the types of relationships/edges/connections in the graph
func EdgeTypes() []string {
	edgeTypes := globalGraph.EdgeTypes()
	sort.Strings(edgeTypes)
	return edgeTypes
}

// NodeTypes returns the types of nodes in the graph
func NodeTypes() []string {
	nodeTypes := globalGraph.NodeTypes()
	sort.Strings(nodeTypes)
	return nodeTypes
}

// GetNode gets a node from the graph
func GetNode(id primitive.TypedID) (*Node, bool) {
	n, ok := globalGraph.GetNode(id)
	if !ok {
		return nil, false
	}
	return &Node{n}, true
}

// RangeNodeTypes iterates over nodes of a given type until the iterator returns false
func RangeNodeTypes(typ primitive.Type, fn func(n *Node) bool) {
	globalGraph.RangeNodeTypes(typ, func(n primitive.Node) bool {
		return fn(&Node{n})
	})
}

// RangeNodes iterates over all nodes until the iterator returns false
func RangeNodes(fn func(n *Node) bool) {
	globalGraph.RangeNodes(func(n primitive.Node) bool {
		return fn(&Node{n})
	})
}

// RangeEdges iterates over all edges/connections until the iterator returns false
func RangeEdges(fn func(e *primitive.Edge) bool) {
	globalGraph.RangeEdges(fn)
}

// RangeEdgeTypes iterates over edges/connections of a given type until the iterator returns false
func RangeEdgeTypes(edgeType primitive.Type, fn func(e *primitive.Edge) bool) {
	globalGraph.RangeEdgeTypes(edgeType, fn)
}

// HasNode returns true if a node with the typed ID exists in the graph
func HasNode(id primitive.TypedID) bool {
	return globalGraph.HasNode(id)
}

// Close closes the global graph instance
func Close() {
	globalGraph.Close()
}
