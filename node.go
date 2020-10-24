package dagger

import (
	"encoding/json"
	"errors"
	"reflect"
)

const (
	ID_KEY   = "_id"
	TYPE_KEY = "_type"
)

// Node is a functional hash table for storing arbitrary data. It is not concurrency safe
type Node map[string]interface{}

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
		return errors.New("dagger: missing node type")
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

// Filter returns a Node of the values that return true from the filter function
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

// Intersection returns the values that exist in both Nodes ref: https://en.wikipedia.org/wiki/Intersection_(set_theory)#:~:text=In%20mathematics%2C%20the%20intersection%20of,that%20also%20belong%20to%20A).
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

// Union returns the all values in both Nodes ref: https://en.wikipedia.org/wiki/Union_(set_theory)
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
		if values, ok := val.(Node); ok {
			return values, true
		}
	}
	return nil, false
}

func (v Node) IsNested(key string) bool {
	_, ok := v.GetNested(key)
	return ok
}

func (v Node) SetNested(key string, values Node) {
	v.Set(key, values)
}

func (v Node) JSON() ([]byte, error) {
	return json.Marshal(v)
}

func (m Node) FromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m Node) UnmarshalFrom(obj json.Marshaler) error {
	bits, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	return m.FromJSON(bits)
}

func (m Node) MarshalTo(obj json.Unmarshaler) error {
	bits, err := m.JSON()
	if err != nil {
		return err
	}
	return obj.UnmarshalJSON(bits)
}
