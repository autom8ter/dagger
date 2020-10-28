package primitive

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

const (
	ID_KEY   = "_id"
	TYPE_KEY = "_type"
)

// Node is a functional hash table for storing arbitrary data. It is not concurrency safe
type Node map[string]interface{}

func NewNode(attributes map[string]interface{}) Node {
	if attributes == nil {
		attributes = map[string]interface{}{}
	}
	if attributes[ID_KEY] == "" {
		attributes[ID_KEY] = UUID()
	}
	if attributes[TYPE_KEY] == "" {
		attributes[TYPE_KEY] = DefaultType
	}
	return attributes
}

func (n Node) Type() string {
	return n.GetString(TYPE_KEY)
}

func (n Node) ID() string {
	return n.GetString(ID_KEY)
}

func (n Node) HasID() bool {
	return n.Exists(ID_KEY)
}

func (n Node) Validate() error {
	if !n.Exists(ID_KEY) {
		return errors.New("dagger: missing node id")
	}
	if !n.Exists(TYPE_KEY) {
		return fmt.Errorf("dagger: missing node type: %v", n)
	}
	return nil
}

func (n Node) SetID(id string) {
	n.Set(ID_KEY, id)
}

func (n Node) SetType(nodeType string) {
	n.Set(TYPE_KEY, nodeType)
}

// Exists returns true if the key exists in the Node
func (m Node) Exists(key string) bool {
	if val, ok := m[key]; ok && val != nil {
		return true
	}
	return false
}

// Set set an entry in the Node
func (m Node) Set(k string, v interface{}) {
	m[k] = v
}

// SetAll set all entries in the Node
func (m Node) SetAll(data map[string]interface{}) {
	if data == nil {
		return
	}
	for k, v := range data {
		m.Set(k, v)
	}
}

// Get gets an entry from the Node by key
func (m Node) Get(key string) interface{} {
	return m[key]
}

// GetString gets an entry from the Node by key
func (m Node) GetString(key string) string {
	if !m.Exists(key) {
		return ""
	}
	return parseString(m[key])
}

func (m Node) GetBool(key string) bool {
	if !m.Exists(key) {
		return false
	}
	return parseBool(m[key])
}

func (m Node) GetInt(key string) int {
	if !m.Exists(key) {
		return 0
	}
	return parseInt(m[key])
}

// Del deletes the entry from the Node by key
func (m Node) Del(key string) {
	delete(m, key)
}

// Range iterates over the Node with the function. If the function returns false, the iteration exits.
func (m Node) Range(iterator func(key string, v interface{}) bool) {
	for k, v := range m {
		if !iterator(k, v) {
			break
		}
	}
}

// Filter returns a Node of the node that return true from the filter function
func (m Node) Filter(filter func(key string, v interface{}) bool) Node {
	data := Node{}
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

// Intersection returns the node that exist in both Nodes ref: https://en.wikipedia.org/wiki/Intersection_(set_theory)#:~:text=In%20mathematics%2C%20the%20intersection%20of,that%20also%20belong%20to%20A).
func (m Node) Intersection(other Node) Node {
	toReturn := Node{}
	m.Range(func(key string, v interface{}) bool {
		if other.Exists(key) {
			toReturn.Set(key, v)
		}
		return true
	})
	return toReturn
}

// Union returns the all node in both Nodes ref: https://en.wikipedia.org/wiki/Union_(set_theory)
func (m Node) Union(other Node) Node {
	toReturn := Node{}
	m.Range(func(k string, v interface{}) bool {
		toReturn.Set(k, v)
		return true
	})
	other.Range(func(k string, v interface{}) bool {
		toReturn.Set(k, v)
		return true
	})
	return toReturn
}

// Copy creates a replica of the Node
func (m Node) Copy() Node {
	copied := Node{}
	if m == nil {
		return copied
	}
	m.Range(func(k string, v interface{}) bool {
		copied.Set(k, v)
		return true
	})
	return copied
}

func (v Node) Equals(other Node) bool {
	return reflect.DeepEqual(v, other)
}

func (v Node) GetNested(key string) (Node, bool) {
	if val, ok := v[key]; ok && val != nil {
		if node, ok := val.(Node); ok {
			return node, true
		}
	}
	return nil, false
}

func (v Node) IsNested(key string) bool {
	_, ok := v.GetNested(key)
	return ok
}

func (v Node) SetNested(key string, node Node) {
	v.Set(key, node)
}

func (v Node) JSON() ([]byte, error) {
	return json.Marshal(v)
}

func (m Node) FromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (n2 Node) Read(p []byte) (n int, err error) {
	bits, err := n2.JSON()
	if err != nil {
		return 0, err
	}
	return copy(p, bits), nil
}

func (n2 Node) Write(p []byte) (n int, err error) {
	if err := n2.FromJSON(p); err != nil {
		return 0, err
	}
	return len(p), nil
}
