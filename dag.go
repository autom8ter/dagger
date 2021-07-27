package dagger

import (
	"fmt"
	"github.com/autom8ter/dagger/ds"
	"github.com/autom8ter/dagger/util"
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

func (g *Graph) SetNode(key ForeignKey, attr Attributes) Node {
	if key.ID() == "" {
		key.SetID(util.UUID())
	}
	n := Node{
		ForeignKey: key,
		Attributes: attr,
	}
	g.nodes.Set(key.Type(), key.ID(), n)
	return n
}

func (g *Graph) GetNode(id TypedID) (Node, bool) {
	val, ok := g.nodes.Get(id.Type(), id.ID())
	if ok {
		n, ok := val.(Node)
		if ok {
			return n, true
		}
	}
	return Node{}, false
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

func (g *Graph) RangeEdges(fn func(e Edge) bool) {
	for _, namespace := range g.edges.Namespaces() {
		g.edges.Range(namespace, func(key string, val interface{}) bool {
			e, ok := val.(Edge)
			if ok {
				if !fn(e) {
					return false
				}
			}
			return true
		})
	}
}

func (g *Graph) RangeEdgeTypes(edgeType Type, fn func(e Edge) bool) {
	g.edges.Range(edgeType.Type(), func(key string, val interface{}) bool {
		e, ok := val.(Edge)
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
			edges.Range(func(e Edge) bool {
				g.DelEdge(e)
				return true
			})
		}
	}
	g.nodes.Delete(id.Type(), id.ID())
}

func (g *Graph) SetEdge(from, to TypedID, node Node) (Edge, error) {
	fromNode, ok := g.GetNode(from)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", from.Type(), from.ID())
	}
	toNode, ok := g.GetNode(to)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", to.Type(), to.ID())
	}
	if !node.ForeignKey.HasID() {
		node.ForeignKey.SetID(util.UUID())
	}
	if !node.ForeignKey.HasType() {
		node.ForeignKey.SetType(DefaultType)
	}

	e := Edge{
		Node: node,
		From: fromNode,
		To:   toNode,
	}

	g.edges.Set(e.ForeignKey.Type(), e.ForeignKey.ID(), e)

	if val, ok := g.edgesFrom.Get(e.From.ForeignKey.Type(), e.From.ForeignKey.ID()); ok && val != nil {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.ForeignKey.Type(), e.From.ForeignKey.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.ForeignKey.Type(), e.From.ForeignKey.ID(), edges)
	}
	if val, ok := g.edgesTo.Get(e.To.ForeignKey.Type(), e.To.ForeignKey.ID()); ok && val != nil {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.ForeignKey.Type(), e.To.ForeignKey.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.ForeignKey.Type(), e.To.ForeignKey.ID(), edges)
	}
	return e, nil
}

func (g *Graph) HasEdge(id TypedID) bool {
	_, ok := g.GetEdge(id)
	return ok
}

func (g *Graph) GetEdge(id TypedID) (Edge, bool) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok {
		e, ok := val.(Edge)
		if ok {
			return e, true
		}
	}
	return Edge{}, false
}

func (g *Graph) DelEdge(id TypedID) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok && val != nil {
		edge := val.(Edge)
		fromVal, ok := g.edgesFrom.Get(edge.From.ForeignKey.Type(), edge.From.ForeignKey.ID())
		if ok && fromVal != nil {
			edges := fromVal.(edgeMap)
			edges.DelEdge(id)
			g.edgesFrom.Set(edge.From.ForeignKey.Type(), edge.From.ForeignKey.ID(), edges)
		}
		toVal, ok := g.edgesTo.Get(edge.To.ForeignKey.Type(), edge.To.ForeignKey.ID())
		if ok && toVal != nil {
			edges := toVal.(edgeMap)
			edges.DelEdge(id)
			g.edgesTo.Set(edge.To.ForeignKey.Type(), edge.To.ForeignKey.ID(), edges)
		}
	}
	g.edges.Delete(id.Type(), id.ID())
}

func (g *Graph) EdgesFrom(edgeType string, id TypedID, fn func(e Edge) bool) {
	val, ok := g.edgesFrom.Get(id.Type(), id.ID())
	if ok {
		if edges, ok := val.(edgeMap); ok {
			edges.RangeType(edgeType, func(e Edge) bool {
				return fn(e)
			})
		}
	}
}

func (g *Graph) EdgesTo(edgeType string, id TypedID, fn func(e Edge) bool) {
	val, ok := g.edgesTo.Get(id.Type(), id.ID())
	if ok {
		if edges, ok := val.(edgeMap); ok {
			edges.RangeType(edgeType, func(e Edge) bool {
				return fn(e)
			})
		}
	}
}

func (g *Graph) Export() *Export {
	exp := &Export{}
	g.RangeNodes(func(n Node) bool {
		exp.Nodes = append(exp.Nodes, n)
		return true
	})
	g.RangeEdges(func(e Edge) bool {
		exp.Edges = append(exp.Edges, e)
		return true
	})
	return exp
}

func (g *Graph) Import(exp *Export) error {
	for _, n := range exp.Nodes {
		g.SetNode(n.ForeignKey, n.Attributes)
	}
	for _, e := range exp.Edges {
		if _, err := g.SetEdge(e.From, e.To, e.Node); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) Close() {
	g.nodes.Close()
	g.edgesTo.Close()
	g.edgesFrom.Close()
	g.edges.Close()
}

func (g *Graph) DFS(edgeType string, rootNode Node, fn func(node Node) bool) {
	var visited = map[ForeignKey]struct{}{}
	stack := ds.NewStack()
	g.dfs(edgeType, rootNode, stack, visited)
	stack.Range(func(element interface{}) bool {
		node := element.(Node)
		if node.XID == rootNode.XID && node.XType == rootNode.XType {
			return true
		}
		return fn(element.(Node))
	})
}

func (g *Graph) dfs(edgeType string, n Node, stack *ds.Stack, visited map[ForeignKey]struct{}) {
	if _, ok := visited[n.ForeignKey]; !ok {
		visited[n.ForeignKey] = struct{}{}
		stack.Push(n)
		g.EdgesFrom(edgeType, n, func(e Edge) bool {
			g.dfs(edgeType, e.To, stack, visited)
			return true
		})
	}
}

