package primitive

import (
	"fmt"
)

// Graph is a concurrency safe, mutable, in-memory directed graph
type Graph struct {
	nodes     *namespacedCache
	edges     *namespacedCache
	edgesFrom *namespacedCache
	edgesTo   *namespacedCache
}

func NewGraph() *Graph {
	return &Graph{
		nodes:     newCache(),
		edges:     newCache(),
		edgesFrom: newCache(),
		edgesTo:   newCache(),
	}
}

func (g *Graph) EdgeTypes() []string {
	return g.edges.Namespaces()
}

func (g *Graph) NodeTypes() []string {
	return g.nodes.Namespaces()
}

func (g *Graph) AddNode(n Node) {
	if n.ID() == "" {
		n.SetID(uuid())
	}
	g.nodes.Set(n.Type(), n.ID(), n)
}

func (g *Graph) AddNodes(nodes ...Node) {
	for _, n := range nodes {
		g.AddNode(n)
	}
}
func (g *Graph) GetNode(id TypedID) (Node, bool) {
	val, ok := g.nodes.Get(id.Type(), id.ID())
	if ok {
		n, ok := val.(Node)
		if ok {
			return n, true
		}
	}
	return nil, false
}

func (g *Graph) RangeNodeTypes(typ Type, fn func(n Node) bool) {
	g.nodes.Range(typ.Type(), func(key string, val interface{}) bool {
		n, ok := val.(Node)
		if ok {
			if !fn(n) {
				return false
			}
		}
		return true
	})
}

func (g *Graph) RangeNodes(fn func(n Node) bool) {
	for _, namespace := range g.nodes.Namespaces() {
		g.nodes.Range(namespace, func(key string, val interface{}) bool {
			n, ok := val.(Node)
			if ok {
				if !fn(n) {
					return false
				}
			}
			return true
		})
	}
}

func (g *Graph) RangeEdges(fn func(e *Edge) bool) {
	for _, namespace := range g.edges.Namespaces() {
		g.edges.Range(namespace, func(key string, val interface{}) bool {
			e, ok := val.(*Edge)
			if ok {
				if !fn(e) {
					return false
				}
			}
			return true
		})
	}
}

func (g *Graph) RangeEdgeTypes(edgeType Type, fn func(e *Edge) bool) {
	g.edges.Range(edgeType.Type(), func(key string, val interface{}) bool {
		e, ok := val.(*Edge)
		if ok {
			if !fn(e) {
				return false
			}
		}
		return true
	})
}

func (g *Graph) HasNode(id TypedID) bool {
	_, ok := g.GetNode(id)
	return ok
}

func (g *Graph) DelNode(id TypedID) {
	if val, ok := g.edgesFrom.Get(id.Type(), id.ID()); ok {
		if val != nil {
			edges := val.(edgeMap)
			edges.Range(func(e *Edge) bool {
				g.DelEdge(e)
				return true
			})
		}
	}
	g.nodes.Delete(id.Type(), id.ID())
}

func (g *Graph) AddEdge(e *Edge) error {
	if e.ID() == "" {
		e.SetID(uuid())
	}
	if err := e.Validate(); err != nil {
		return err
	}
	if !g.HasNode(e.From) {
		return fmt.Errorf("node %s does not exist", e.From.Type())
	}
	if !g.HasNode(e.To) {
		return fmt.Errorf("node does not exist")
	}
	g.edges.Set(e.Type(), e.ID(), e)
	if val, ok := g.edgesFrom.Get(e.From.Type(), e.From.ID()); ok {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.Type(), e.From.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.Type(), e.From.ID(), edges)
	}
	if val, ok := g.edgesTo.Get(e.To.Type(), e.To.ID()); ok {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.Type(), e.To.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.Type(), e.To.ID(), edges)
	}
	return nil
}

func (g *Graph) AddEdges(edges ...*Edge) error {
	for _, e := range edges {
		if err := g.AddEdge(e); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) HasEdge(id TypedID) bool {
	_, ok := g.GetEdge(id)
	return ok
}

func (g *Graph) GetEdge(id TypedID) (*Edge, bool) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok {
		e, ok := val.(*Edge)
		if ok {
			return e, true
		}
	}
	return nil, false
}

func (g *Graph) DelEdge(id TypedID) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok && val != nil {
		edge := val.(*Edge)
		fromVal, ok := g.edgesFrom.Get(edge.From.Type(), edge.From.ID())
		if ok && fromVal != nil {
			edges := fromVal.(edgeMap)
			edges.DelEdge(id)
			g.edgesFrom.Set(edge.From.Type(), edge.From.ID(), edges)
		}
		toVal, ok := g.edgesTo.Get(edge.To.Type(), edge.To.ID())
		if ok && toVal != nil {
			edges := toVal.(edgeMap)
			edges.DelEdge(id)
			g.edgesTo.Set(edge.To.Type(), edge.To.ID(), edges)
		}
	}
	g.edges.Delete(id.Type(), id.ID())
}

func (g *Graph) EdgesFrom(id TypedID, fn func(e *Edge) bool) {
	val, ok := g.edgesFrom.Get(id.Type(), id.ID())
	if ok {
		if edges, ok := val.(edgeMap); ok {
			edges.Range(func(e *Edge) bool {
				return fn(e)
			})
		}
	}
}

func (g *Graph) EdgesTo(id TypedID, fn func(e *Edge) bool) {
	val, ok := g.edgesTo.Get(id.Type(), id.ID())
	if ok {
		if edges, ok := val.(edgeMap); ok {
			edges.Range(func(e *Edge) bool {
				return fn(e)
			})
		}
	}
}

func (g *Graph) Raw() map[string]map[string]map[string]interface{} {
	raw := map[string]map[string]map[string]interface{}{}
	raw["edgesFrom"] = g.edgesFrom.Raw()
	raw["edgesTo"] = g.edgesTo.Raw()
	raw["edges"] = g.edges.Raw()
	raw["nodes"] = g.nodes.Raw()
	return raw
}

func (g *Graph) FromRaw(data map[string]map[string]map[string]interface{}) {
	for k, v := range data {
		if k == "edgesFrom" {
			g.edgesFrom.FromRaw(v)
		}
		if k == "edgesTo" {
			g.edgesTo.FromRaw(v)
		}
		if k == "edges" {
			g.edges.FromRaw(v)
		}
		if k == "nodes" {
			g.nodes.FromRaw(v)
		}
	}
}

func (g *Graph) Close() {
	g.nodes.Close()
	g.edgesTo.Close()
	g.edgesFrom.Close()
	g.edges.Close()
}
