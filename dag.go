package dagger

import (
	"fmt"
	"github.com/autom8ter/dagger/constants"
	"github.com/autom8ter/dagger/ds"
	"github.com/autom8ter/dagger/util"
)

// Graph is a concurrency safe, mutable, in-memory directed graph
type Graph struct {
	nodes     ds.NamespacedCache
	edges     ds.NamespacedCache
	edgesFrom ds.NamespacedCache
	edgesTo   ds.NamespacedCache
}

func NewGraph() *Graph {
	return &Graph{
		nodes:     ds.NewCache(),
		edges:     ds.NewCache(),
		edgesFrom: ds.NewCache(),
		edgesTo:   ds.NewCache(),
	}
}

func (g *Graph) EdgeTypes() []string {
	return g.edges.Namespaces()
}

func (g *Graph) NodeTypes() []string {
	return g.nodes.Namespaces()
}

func (g *Graph) SetNode(key Path, attr Attributes) Node {
	if key.ID() == "" {
		key.SetID(util.UUID())
	}
	n := Node{
		Path:       key,
		Attributes: attr,
	}
	g.nodes.Set(key.Type(), key.ID(), n)
	return n
}

func (g *Graph) GetNode(id Path) (Node, bool) {
	val, ok := g.nodes.Get(id.Type(), id.ID())
	if ok {
		n, ok := val.(Node)
		if ok {
			return n, true
		}
	}
	return Node{}, false
}

func (g *Graph) RangeNodeTypes(typ string, fn func(n Node) bool) {
	g.nodes.Range(typ, func(key string, val interface{}) bool {
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

func (g *Graph) RangeEdgeTypes(edgeType string, fn func(e Edge) bool) {
	g.edges.Range(edgeType, func(key string, val interface{}) bool {
		e, ok := val.(Edge)
		if ok {
			if !fn(e) {
				return false
			}
		}
		return true
	})
}

func (g *Graph) HasNode(id Path) bool {
	_, ok := g.GetNode(id)
	return ok
}

func (g *Graph) DelNode(id Path) {
	if val, ok := g.edgesFrom.Get(id.Type(), id.ID()); ok {
		if val != nil {
			edges := val.(edgeMap)
			edges.Range(func(e Edge) bool {
				g.DelEdge(e.Path)
				return true
			})
		}
	}
	g.nodes.Delete(id.Type(), id.ID())
}

func (g *Graph) SetEdge(from, to Path, node Node) (Edge, error) {
	fromNode, ok := g.GetNode(from)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", from.Type(), from.ID())
	}
	toNode, ok := g.GetNode(to)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", to.Type(), to.ID())
	}
	if !node.Path.HasID() {
		node.Path.SetID(util.UUID())
	}
	if !node.Path.HasType() {
		node.Path.SetType(constants.DefaultType)
	}

	e := Edge{
		Node: node,
		From: fromNode.Path,
		To:   toNode.Path,
	}

	g.edges.Set(e.Path.Type(), e.Path.ID(), e)

	if val, ok := g.edgesFrom.Get(e.From.Type(), e.From.ID()); ok && val != nil {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.Type(), e.From.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesFrom.Set(e.From.Type(), e.From.ID(), edges)
	}
	if val, ok := g.edgesTo.Get(e.To.Type(), e.To.ID()); ok && val != nil {
		edges := val.(edgeMap)
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.Type(), e.To.ID(), edges)
	} else {
		edges := edgeMap{}
		edges.AddEdge(e)
		g.edgesTo.Set(e.To.Type(), e.To.ID(), edges)
	}
	return e, nil
}

func (g *Graph) HasEdge(id Path) bool {
	_, ok := g.GetEdge(id)
	return ok
}

func (g *Graph) GetEdge(id Path) (Edge, bool) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok {
		e, ok := val.(Edge)
		if ok {
			return e, true
		}
	}
	return Edge{}, false
}

func (g *Graph) DelEdge(id Path) {
	val, ok := g.edges.Get(id.Type(), id.ID())
	if ok && val != nil {
		edge := val.(Edge)
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

func (g *Graph) EdgesFrom(edgeType string, id Path, fn func(e Edge) bool) {
	val, ok := g.edgesFrom.Get(id.Type(), id.ID())
	if ok {
		if edges, ok := val.(edgeMap); ok {
			edges.RangeType(edgeType, func(e Edge) bool {
				return fn(e)
			})
		}
	}
}

func (g *Graph) EdgesTo(edgeType string, id Path, fn func(e Edge) bool) {
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
		g.SetNode(n.Path, n.Attributes)
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

func (g *Graph) DFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	stack := ds.NewStack()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.dfs(edgeType, root, stack, visited)
	stack.Range(func(element interface{}) bool {
		path := element.(Node)
		if path.XID == rootNode.XID && path.XType == rootNode.XType {
			return true
		}
		return fn(path)
	})
}

func (g *Graph) dfs(edgeType string, n Node, stack ds.Stack, visited map[Path]struct{}) {
	if _, ok := visited[n.Path]; !ok {
		visited[n.Path] = struct{}{}
		stack.Push(n)
		g.EdgesFrom(edgeType, n.Path, func(e Edge) bool {
			to, ok := g.GetNode(e.To)
			if ok {
				g.dfs(edgeType, to, stack, visited)
			}
			return true
		})
	}
}


func (g *Graph) BFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	stack := ds.NewStack()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.dfs(edgeType, root, stack, visited)
	stack.Range(func(element interface{}) bool {
		path := element.(Node)
		if path.XID == rootNode.XID && path.XType == rootNode.XType {
			return true
		}
		return fn(path)
	})
}

func (g *Graph) bfs(edgeType string, n Node, stack ds.Stack, visited map[Path]struct{}) {
	if _, ok := visited[n.Path]; !ok {
		visited[n.Path] = struct{}{}
		stack.Push(n)
		g.EdgesFrom(edgeType, n.Path, func(e Edge) bool {
			to, ok := g.GetNode(e.To)
			if ok {
				g.dfs(edgeType, to, stack, visited)
			}
			return true
		})
	}
}
