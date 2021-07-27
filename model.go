package dagger

import (
	"encoding/json"
	"errors"
	"github.com/autom8ter/dagger/util"
)

const DefaultType = "default"

const AnyType = "*"

type ID interface {
	ID() string
}

type Type interface {
	Type() string
}

type TypedID interface {
	ID
	Type
}

// ForeignKey satisfies primitive.TypedID interface
type ForeignKey struct {
	XID   string `json:"xid"`
	XType string `json:"xtype"`
}

func (n ForeignKey) HasID() bool {
	return n.XID != ""
}

func (n ForeignKey) HasType() bool {
	return n.XType != ""
}

func (n ForeignKey) SetID(id string) {
	n.XID = id
}

func (n ForeignKey) SetType(typ string) {
	n.XType = typ
}

func (f ForeignKey) ID() string {
	return f.XID
}

func (f ForeignKey) Type() string {
	return f.XType
}

func (n ForeignKey) Validate() error {
	if !n.HasID() {
		return errors.New("dagger: missing node id")
	}
	if !n.HasType() {
		return errors.New("dagger: missing node type")
	}
	return nil
}

// Node is a functional hash table for storing arbitrary data. It is not concurrency safe
type Node struct {
	ForeignKey
	Attributes map[string]interface{}
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

type Attributes map[string]interface{}

func (m Attributes) Exists(key string) bool {
	if val, ok := m[key]; ok && val != nil {
		return true
	}
	return false
}

// Set set an entry in the Node
func (m Attributes) Set(k string, v interface{}) {
	m[k] = v
}

// SetAll set all entries in the Node
func (m Attributes) SetAll(data map[string]interface{}) {
	if data == nil {
		return
	}
	for k, v := range data {
		m.Set(k, v)
	}
}

// Get gets an entry from the Attributes by key
func (m Attributes) Get(key string) interface{} {
	return m[key]
}

// GetString gets an entry from the Attributes by key
func (m Attributes) GetString(key string) string {
	if !m.Exists(key) {
		return ""
	}
	return util.ParseString(m[key])
}

func (m Attributes) GetBool(key string) bool {
	if !m.Exists(key) {
		return false
	}
	return util.ParseBool(m[key])
}

func (m Attributes) GetInt(key string) int {
	if !m.Exists(key) {
		return 0
	}
	return util.ParseInt(m[key])
}

// Del deletes the entry from the Attributes by key
func (m Attributes) Del(key string) {
	delete(m, key)
}

// Range iterates over the Attributes with the function. If the function returns false, the iteration exits.
func (m Attributes) Range(iterator func(key string, v interface{}) bool) {
	for k, v := range m {
		if !iterator(k, v) {
			break
		}
	}
}

// Filter returns a Attributes of the node that return true from the filter function
func (m Attributes) Filter(filter func(key string, v interface{}) bool) Attributes {
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

// Edge is a relationship between two nodes
type Edge struct {
	// An edge implements Node because it has an Identifier and attributes
	Node `json:"node"`
	// From returns the root node of the edge
	From Node `json:"from"`
	// To returns the target node of the edge
	To Node `json:"to"`
}

func (e Edge) JSON() ([]byte, error) {
	return json.Marshal(&e)
}

// edgeMap is a map of edges. edgeMap are not concurrency safe.
type edgeMap map[string]map[string]Edge

func (e edgeMap) Types() []string {
	var typs []string
	for t, _ := range e {
		typs = append(typs, t)
	}
	return typs
}

// RangeType executes the function over a list of edges with the given type. If the function returns false, the iteration stops.
func (e edgeMap) RangeType(typ string, fn func(e Edge) bool) {
	if typ == AnyType {
		for _, edges := range e {
			for _, edge := range edges {
				if !fn(edge) {
					break
				}
			}
		}
	} else {
		if e[typ] == nil {
			return
		}
		for _, e := range e[typ] {
			if !fn(e) {
				break
			}
		}
	}
}

// Range executes the function over every edge. If the function returns false, the iteration stops.
func (e edgeMap) Range(fn func(e Edge) bool) {
	for _, m := range e {
		for _, e := range m {
			if !fn(e) {
				break
			}
		}
	}
}

// Filter executes the function over every edge. If the function returns true, the edges will be added to the returned array of edges.
func (e edgeMap) Filter(fn func(e Edge) bool) []Edge {
	var edges []Edge
	for _, m := range e {
		for _, e := range m {
			if fn(e) {
				edges = append(edges, e)
			}
		}
	}
	return edges
}

// FilterType executes the function over every edge of the given type. If the function returns true, the edges will be added to the returned array of edges.
func (e edgeMap) FilterType(typ string, fn func(e Edge) bool) []Edge {
	var edges []Edge
	e.RangeType(typ, func(e Edge) bool {
		if fn(e) {
			edges = append(edges, e)
		}
		return true
	})
	return edges
}

// DelEdge deletes the edge
func (e edgeMap) DelEdge(id TypedID) {
	if _, ok := e[id.Type()]; !ok {
		return
	}
	delete(e[id.Type()], id.ID())
}

// AddEdge adds the edge to the map
func (e edgeMap) AddEdge(edge Edge) {
	if _, ok := e[edge.ForeignKey.Type()]; !ok {
		e[edge.ForeignKey.Type()] = map[string]Edge{
			edge.ForeignKey.ID(): edge,
		}
	} else {
		e[edge.ForeignKey.Type()][edge.ForeignKey.ID()] = edge
	}
}

// HasEdge returns true if the edge exists
func (e edgeMap) HasEdge(id TypedID) bool {
	_, ok := e.GetEdge(id)
	return ok
}

// GetEdge gets an edge by id
func (e edgeMap) GetEdge(id TypedID) (Edge, bool) {
	if _, ok := e[id.Type()]; !ok {
		return Edge{}, false
	}
	if e, ok := e[id.Type()][id.ID()]; ok {
		return e, true
	}
	return Edge{}, false
}

// Len returns the number of edges of the given type
func (e edgeMap) Len(typ Type) int {
	if rels, ok := e[typ.Type()]; ok {
		return len(rels)
	}
	return 0
}

type Export struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
