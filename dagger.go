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

//
//func DelNode(id primitive.TypedID) {
//	globalGraph.DelNode(id)
//}
//
//func AddEdge(e *primitive.Edge) error {
//	return globalGraph.AddEdge(e)
//}
//
//func AddEdges(edges ...*primitive.Edge) error {
//	return globalGraph.AddEdges(edges...)
//}
//
//func HasEdge(id primitive.TypedID) bool {
//	return globalGraph.HasEdge(id)
//}
//
//func GetEdge(id primitive.TypedID) (*primitive.Edge, bool) {
//	return globalGraph.GetEdge(id)
//}
//
//func DelEdge(id primitive.TypedID) {
//	globalGraph.DelEdge(id)
//}
//
//func EdgesFrom(id primitive.TypedID, fn func(e *primitive.Edge) bool) {
//	globalGraph.EdgesFrom(id, fn)
//}
//
//func EdgesTo(id primitive.TypedID, fn func(e *primitive.Edge) bool) {
//	globalGraph.EdgesTo(id, fn)
//}

func Close() {
	globalGraph.Close()
}
