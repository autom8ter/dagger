package dagger

import (
	"github.com/autom8ter/dagger/util"
)

// Edge is a relationship between two nodes
type Edge struct {
	// An edge implements Node because it has an Identifier and attributes
	Node `json:"node"`
	// From returns the root node of the edge
	From Path `json:"from"`
	// To returns the target node of the edge
	To Path `json:"to"`
}

// Path satisfies primitive.Path interface
type Path struct {
	XID   string `json:"xid"`
	XType string `json:"xtype"`
}

// Node is a single entity in the graph.
type Node struct {
	Path       `json:"path"`
	Attributes Attributes `json:"attributes"`
}

type Attributes map[string]interface{}

// Exists checks for the existance of a key
func (m Attributes) Exists(key string) bool {
	if m == nil {
		m = map[string]interface{}{}
	}
	if val, ok := m[key]; ok && val != nil {
		return true
	}
	return false
}

// Set set an entry in the Node
func (m Attributes) Set(k string, v interface{}) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m[k] = v
}

// SetAll set all entries in the Node
func (m Attributes) SetAll(data map[string]interface{}) {
	if m == nil {
		m = map[string]interface{}{}
	}
	for k, v := range data {
		m.Set(k, v)
	}
}

// Get gets an entry from the Attributes by key
func (m Attributes) Get(key string) interface{} {
	if m == nil {
		m = map[string]interface{}{}
	}
	return m[key]
}

// GetString gets an entry from the Attributes by key
func (m Attributes) GetString(key string) string {
	if m == nil {
		m = map[string]interface{}{}
	}
	if !m.Exists(key) {
		return ""
	}
	return util.ParseString(m[key])
}

// GetBool gets an entry from the Attributes by key
func (m Attributes) GetBool(key string) bool {
	if m == nil {
		m = map[string]interface{}{}
	}
	if !m.Exists(key) {
		return false
	}
	return util.ParseBool(m[key])
}

// GetInt gets an entry from the Attributes by key
func (m Attributes) GetInt(key string) int {
	if m == nil {
		m = map[string]interface{}{}
	}
	if !m.Exists(key) {
		return 0
	}
	return util.ParseInt(m[key])
}

// Del deletes the entry from the Attributes by key
func (m Attributes) Del(key string) {
	if m == nil {
		m = map[string]interface{}{}
	}
	delete(m, key)
}

// Range iterates over the Attributes with the function. If the function returns false, the iteration exits.
func (m Attributes) Range(iterator func(key string, v interface{}) bool) {
	if m == nil {
		m = map[string]interface{}{}
	}
	for k, v := range m {
		if !iterator(k, v) {
			break
		}
	}
}

// Filter returns a Attributes of the node that return true from the filter function
func (m Attributes) Filter(filter func(key string, v interface{}) bool) Attributes {
	if m == nil {
		m = map[string]interface{}{}
	}
	data := Attributes{}
	if m == nil {
		return data
	}
	m.Range(func(key string, v interface{}) bool {
		if filter(key, v) {
			data.Set(key, v)
		}
		return true
	})
	return data
}

// Copy creates a replica of the Node
func (m Attributes) Copy() Attributes {
	if m == nil {
		m = map[string]interface{}{}
	}
	copied := Attributes{}
	if m == nil {
		return copied
	}
	m.Range(func(k string, v interface{}) bool {
		copied.Set(k, v)
		return true
	})
	return copied
}

// Export contains an array of nodes and their corresponding edges
type Export struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
