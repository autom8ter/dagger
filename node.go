package dagger

import "github.com/autom8ter/dagger/primitive"

func NewNode(nodeType, id string, attributes map[string]interface{}) *Node {
	data := primitive.NewNode(nodeType, id)
	data.SetAll(attributes)
	if !globalGraph.HasNode(data) || !data.HasID() {
		globalGraph.AddNode(data)
		return &Node{data}
	}
	return &Node{TypedID: data}
}

type Node struct {
	primitive.TypedID
}

func (n *Node) EdgesFrom(fn func(e *primitive.Edge) bool) {
	globalGraph.EdgesFrom(n, fn)
}

func (n *Node) EdgesTo(fn func(e *primitive.Edge) bool) {
	globalGraph.EdgesTo(n, fn)
}

func (n *Node) Remove() {
	globalGraph.DelNode(n)
}

func (n *Node) load() primitive.Node {
	node, ok := globalGraph.GetNode(n)
	if !ok {
		globalGraph.AddNode(primitive.NewNode(n.Type(), n.ID()))
		node, ok = globalGraph.GetNode(n)
	}
	return node
}

func (n *Node) Connect(node *Node, relationship string) error {
	return globalGraph.AddEdge(&primitive.Edge{
		Node: primitive.NewNode(relationship, ""),
		From: n.load(),
		To:   node.load(),
	})
}

func (n *Node) Patch(data map[string]interface{}) {
	node := n.load()
	node.SetAll(data)
	globalGraph.AddNode(node)
}

func (n *Node) Range(fn func(key string, value interface{}) bool ) {
	node := n.load()
	node.Range(fn)
}

func (n *Node) JSON() ([]byte, error) {
	return n.load().JSON()
}
