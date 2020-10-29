package dagger

import (
	"github.com/autom8ter/dagger/primitive"
)

// Edge is an edge in the directed graph. It represents a relationship between two nodes.
type Edge struct {
	primitive.TypedID
}

// NewEdge creates a new edge node in the global, in-memory graph.
func NewEdge(relationship string, from, to *Node, mutual bool) (*Edge, error) {
	return from.Connect(to, relationship, mutual)
}

func edgeFrom(edge *primitive.Edge) (*Edge, error) {
	if !globalGraph.HasEdge(edge) || !edge.HasID() {
		if err := globalGraph.AddEdge(edge); err != nil {
			return nil, err
		}
		return &Edge{edge}, nil
	}
	return &Edge{edge}, nil
}

func (e *Edge) load() *primitive.Edge {
	edge, ok := globalGraph.GetEdge(e)
	if !ok {
		return &primitive.Edge{
			Node: primitive.Node{},
			From: primitive.Node{},
			To:   primitive.Node{},
		}
	}
	return edge
}

// From returns the node that points to the node returned by To()
func (e *Edge) From() *Node {
	return nodeFrom(e.load().From)
}

// To returns the node that is being pointed to by From()
func (e *Edge) To() *Node {
	return nodeFrom(e.load().To)
}

// Patch patches the edge attributes with the given data
func (e *Edge) Patch(data map[string]interface{}) {
	edge := e.load()
	edge.SetAll(data)
	globalGraph.AddEdge(edge)
}

// Range iterates over the edges attributes until the iterator returns false
func (e *Edge) Range(fn func(key string, value interface{}) bool) {
	edge := e.load()
	edge.Range(fn)
}

func (e *Edge) Node() *Node {
	edge := e.load()
	return nodeFrom(edge.Node)
}

// GetString gets a string value from the edges attributes(if it exists)
func (e *Edge) GetString(key string) string {
	edge := e.load()
	return edge.GetString(key)
}

// GetInt gets an int value from the edges attributes(if it exists)
func (e *Edge) GetInt(key string) int {
	edge := e.load()
	return edge.GetInt(key)
}

// GetBool gets a bool value from the edges attributes(if it exists)
func (e *Edge) GetBool(key string) bool {
	edge := e.load()
	return edge.GetBool(key)
}

// Get gets an empty interface value(any value type) from the edges attributes(if it exists)
func (e *Edge) Get(key string) interface{} {
	edge := e.load()
	return edge.Get(key)
}

// Del deletes the entry from the edge by key
func (e *Edge) Del(key string) {
	edge := e.load()
	edge.Del(key)
}

// JSON returns the edge as JSON bytes
func (e *Edge) JSON() ([]byte, error) {
	return e.load().JSON()
}

// FromJSON encodes the edge with the given JSON bytes
func (e *Edge) FromJSON(bits []byte) error {
	edge := e.load()
	return edge.FromJSON(bits)
}
