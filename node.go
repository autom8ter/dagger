package dagger

import (
	"fmt"
	"github.com/autom8ter/dagger/primitive"
)

// NewNode creates a new node in the global, in-memory graph.
// If an id is not provided, a random uuid will be assigned.
func NewNode(nodeType, id string, attributes map[string]interface{}) *Node {
	data := primitive.NewNode(nodeType, id)
	data.SetAll(attributes)
	return nodeFrom(data)
}

func nodeFrom(node primitive.Node) *Node {
	if !globalGraph.HasNode(node) || !node.HasID() {
		globalGraph.AddNode(node)
		return &Node{node}
	}
	return &Node{TypedID: node}
}

// Node is the most basic element in the graph. Node's may be connected with one another via edges to represent relationships
type Node struct {
	primitive.TypedID
}

// EdgesFrom returns connections/edges that stem from the node/vertex
func (n *Node) EdgesFrom(fn func(edge *Edge) bool) {
	globalGraph.EdgesFrom(n, func(e *primitive.Edge) bool {
		this, err := edgeFrom(e)
		if err != nil {
			panic(err)
		}
		return fn(this)
	})
}

func (n *Node) load() primitive.Node {
	node, ok := globalGraph.GetNode(n)
	if !ok {
		globalGraph.AddNode(primitive.NewNode(n.Type(), n.ID()))
		node, ok = globalGraph.GetNode(n)
	}
	return node
}

// EdgesTo returns connections/edges that point toward the node/vertex
func (n *Node) EdgesTo(fn func(e *Edge) bool) {
	globalGraph.EdgesTo(n, func(e *primitive.Edge) bool {
		this, err := edgeFrom(e)
		if err != nil {
			panic(err)
		}
		return fn(this)
	})
}

// Remove permenently removes the node from the graph
func (n *Node) Remove() {
	globalGraph.DelNode(n)
}

// Connect creates a connection/edge between the two nodes with the given relationship type
// if mutual = true, the connection is doubly linked - (facebook is mutual, instagram is not)
func (n *Node) Connect(nodeID primitive.TypedID, relationship string, mutual bool) error {
	node, ok := GetNode(nodeID)
	if !ok {
		return fmt.Errorf("node: %s %s does not exist", nodeID.Type(), nodeID.ID())
	}
	if !mutual {
		return globalGraph.AddEdge(&primitive.Edge{
			Node: primitive.NewNode(relationship, ""),
			From: n.load(),
			To:   node.load(),
		})
	}
	if err := globalGraph.AddEdge(&primitive.Edge{
		Node: primitive.NewNode(relationship, ""),
		From: n.load(),
		To:   node.load(),
	}); err != nil {
		return err
	}
	if err := globalGraph.AddEdge(&primitive.Edge{
		Node: primitive.NewNode(relationship, ""),
		From: node.load(),
		To:   n.load(),
	}); err != nil {
		return err
	}
	return nil
}

// Patch patches the node attributes with the given data
func (n *Node) Patch(data map[string]interface{}) {
	node := n.load()
	node.SetAll(data)
	globalGraph.AddNode(node)
}

// Range iterates over the nodes attributes until the iterator returns false
func (n *Node) Range(fn func(key string, value interface{}) bool) {
	node := n.load()
	node.Range(fn)
}

// GetString gets a string value from the nodes attributes(if it exists)
func (n *Node) GetString(key string) string {
	node := n.load()
	return node.GetString(key)
}

// GetInt gets an int value from the nodes attributes(if it exists)
func (n *Node) GetInt(key string) int {
	node := n.load()
	return node.GetInt(key)
}

// GetBool gets a bool value from the nodes attributes(if it exists)
func (n *Node) GetBool(key string) bool {
	node := n.load()
	return node.GetBool(key)
}

// Get gets an empty interface value(any value type) from the nodes attributes(if it exists)
func (n *Node) Get(key string) interface{} {
	node := n.load()
	return node.Get(key)
}

// Del deletes the entry from the Node by key
func (n *Node) Del(key string) {
	node := n.load()
	node.Del(key)
}

// JSON returns the node as JSON bytes
func (n *Node) JSON() ([]byte, error) {
	return n.load().JSON()
}

// FromJSON encodes the node with the given JSON bytes
func (n *Node) FromJSON(bits []byte) error {
	node := n.load()
	return node.FromJSON(bits)
}

// Raw returns the underlying map[string]interface{}. The map should be treated as readonly.
func (n *Node) Raw() map[string]interface{} {
	return n.load()
}
