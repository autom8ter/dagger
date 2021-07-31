package dagger

import (
	"fmt"
	"github.com/autom8ter/dagger/constants"
	"github.com/autom8ter/dagger/driver"
	"github.com/autom8ter/dagger/util"
)

// Graph is a concurrency safe, mutable, in-memory directed graph
type Graph struct {
	nodes     driver.Index
	edges     driver.Index
	edgesFrom driver.Index
	edgesTo   driver.Index
}

// NewGraph creates a new in-memory Graph instance
func NewGraph() *Graph {
	return &Graph{
		nodes:     driver.NewInMemIndex(),
		edges:     driver.NewInMemIndex(),
		edgesFrom: driver.NewInMemIndex(),
		edgesTo:   driver.NewInMemIndex(),
	}
}

// CustomGraph creates a new Graph instance with the given Index interface implementations
func CustomGraph(nodes, edges, edgesFrom, edgesTo driver.Index) *Graph {
	return &Graph{
		nodes:     nodes,
		edges:     edges,
		edgesFrom: edgesFrom,
		edgesTo:   edgesTo,
	}
}

// EdgeTypes returns all edge types in the graph
func (g *Graph) EdgeTypes() []string {
	return g.edges.Namespaces()
}

// EdgeTypes returns all node types in the graph
func (g *Graph) NodeTypes() []string {
	return g.nodes.Namespaces()
}

// SetNode sets a node in the graph
func (g *Graph) SetNode(key Path, attr Attributes) Node {
	if key.XID == "" {
		key = Path{
			XID:   util.UUID(),
			XType: key.XType,
		}
	}
	if key.XType == "" {
		key = Path{
			XID:   key.XID,
			XType: constants.DefaultType,
		}
	}
	n := Node{
		Path:       key,
		Attributes: attr,
	}
	g.nodes.Set(key.XType, key.XID, n)
	return n
}

// GetNode gets an existing node in the graph
func (g *Graph) GetNode(id Path) (Node, bool) {
	val, ok := g.nodes.Get(id.XType, id.XID)
	if ok {
		n, ok := val.(Node)
		if ok {
			return n, true
		}
	}
	return Node{}, false
}

// RangeNodes ranges over nodes of the given type until the given function returns false
func (g *Graph) RangeNodes(nodeType string, fn func(n Node) bool) {
	g.nodes.Range(nodeType, func(key string, val interface{}) bool {
		n, ok := val.(Node)
		if ok {
			if !fn(n) {
				return false
			}
		}
		return true
	})
}

// RangeEdges ranges over edges until the given function returns false
func (g *Graph) RangeEdges(edgeType string, fn func(e Edge) bool) {
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
	if val, ok := g.edgesFrom.Get(id.XType, id.XID); ok {
		if val != nil {
			edges := val.(driver.Index)
			edges.Range("*", func(key string, val interface{}) bool {
				g.DelEdge(val.(Edge).Path)
				return true
			})
		}
	}
	if val, ok := g.edgesTo.Get(id.XType, id.XID); ok {
		if val != nil {
			edges := val.(driver.Index)
			edges.Range("*", func(key string, val interface{}) bool {
				g.DelEdge(val.(Edge).Path)
				return true
			})
		}
	}
	g.nodes.Delete(id.XType, id.XID)
}

func (g *Graph) SetEdge(from, to Path, node Node) (Edge, error) {
	fromNode, ok := g.GetNode(from)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", from.XType, from.XID)
	}
	toNode, ok := g.GetNode(to)
	if !ok {
		return Edge{}, fmt.Errorf("node %s.%s does not exist", to.XType, to.XID)
	}
	if node.Path.XID == "" {
		node.Path = Path{
			XID:   util.UUID(),
			XType: node.Path.XType,
		}
	}
	if node.Path.XType == "" {
		node.Path = Path{
			XID:   node.Path.XID,
			XType: constants.DefaultType,
		}
	}
	e := Edge{
		Node: node,
		From: fromNode.Path,
		To:   toNode.Path,
	}

	g.edges.Set(e.Path.XType, e.Path.XID, e)

	if val, ok := g.edgesFrom.Get(e.From.XType, e.From.XID); ok && val != nil {
		edges := val.(driver.Index)
		edges.Set(e.XType, e.XID, e)
		g.edgesFrom.Set(e.From.XType, e.From.XID, edges)
	} else {
		edges := driver.NewInMemIndex()
		edges.Set(e.XType, e.XID, e)
		g.edgesFrom.Set(e.From.XType, e.From.XID, edges)
	}
	if val, ok := g.edgesTo.Get(e.To.XType, e.To.XID); ok && val != nil {
		edges := val.(driver.Index)
		edges.Set(e.XType, e.XID, e)
		g.edgesTo.Set(e.To.XType, e.To.XID, edges)
	} else {
		edges := driver.NewInMemIndex()
		edges.Set(e.XType, e.XID, e)
		g.edgesTo.Set(e.To.XType, e.To.XID, edges)
	}
	return e, nil
}

func (g *Graph) HasEdge(id Path) bool {
	_, ok := g.GetEdge(id)
	return ok
}

func (g *Graph) GetEdge(id Path) (Edge, bool) {
	val, ok := g.edges.Get(id.XType, id.XID)
	if ok {
		e, ok := val.(Edge)
		if ok {
			return e, true
		}
	}
	return Edge{}, false
}

func (g *Graph) DelEdge(id Path) {
	val, ok := g.edges.Get(id.XType, id.XID)
	if ok && val != nil {
		edge := val.(Edge)
		fromVal, ok := g.edgesFrom.Get(edge.From.XType, edge.From.XID)
		if ok && fromVal != nil {
			edges := fromVal.(driver.Index)
			edges.Delete(id.XType, id.XID)
			g.edgesFrom.Set(edge.From.XType, edge.From.XID, edges)
		}
		toVal, ok := g.edgesTo.Get(edge.To.XType, edge.To.XID)
		if ok && toVal != nil {
			edges := toVal.(driver.Index)
			edges.Delete(id.XType, id.XID)
			g.edgesTo.Set(edge.To.XType, edge.To.XID, edges)
		}
	}
	g.edges.Delete(id.XType, id.XID)
}

func (g *Graph) RangeEdgesFrom(edgeType string, id Path, fn func(e Edge) bool) {
	val, ok := g.edgesFrom.Get(id.XType, id.XID)
	if ok {
		if edges, ok := val.(driver.Index); ok {
			edges.Range(edgeType, func(key string, val interface{}) bool {
				return fn(val.(Edge))
			})
		}
	}
}

func (g *Graph) RangeEdgesTo(edgeType string, id Path, fn func(e Edge) bool) {
	val, ok := g.edgesTo.Get(id.XType, id.XID)
	if ok {
		if edges, ok := val.(driver.Index); ok {
			edges.Range(edgeType, func(key string, val interface{}) bool {
				return fn(val.(Edge))
			})
		}
	}
}

func (g *Graph) Export() *Export {
	exp := &Export{}
	g.RangeNodes("*", func(n Node) bool {
		exp.Nodes = append(exp.Nodes, n)
		return true
	})
	g.RangeEdges("*", func(e Edge) bool {
		exp.Edges = append(exp.Edges, e)
		return true
	})
	return exp
}

// Import imports the export instance into the graph
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

// Close closes the graph
func (g *Graph) Close() {
	g.nodes.Close()
	g.edgesTo.Close()
	g.edgesFrom.Close()
	g.edges.Close()
}

// DFS executes a depth first search with the rootNode and edge type
func (g *Graph) DFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	stack := driver.NewInMemStack()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.dfs(false, edgeType, &root, stack, visited)
	stack.Range(func(element interface{}) bool {
		path := element.(*Node)
		if path.Path != rootNode {
			return fn(*path)
		}
		return true
	})
}

// ReverseDFS executes a reverse depth first search with the rootNode and edge type
func (g *Graph) ReverseDFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	stack := driver.NewInMemStack()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.dfs(true, edgeType, &root, stack, visited)
	stack.Range(func(element interface{}) bool {
		path := element.(*Node)
		if path.Path != rootNode {
			return fn(*path)
		}
		return true
	})
}

func (g *Graph) dfs(reverse bool, edgeType string, n *Node, stack driver.Stack, visited map[Path]struct{}) {
	if _, ok := visited[n.Path]; !ok {
		visited[n.Path] = struct{}{}
		stack.Push(n)
		if reverse {
			g.RangeEdgesTo(edgeType, n.Path, func(e Edge) bool {
				to, _ := g.GetNode(e.From)
				g.dfs(reverse, edgeType, &to, stack, visited)
				return true
			})
		} else {
			g.RangeEdgesFrom(edgeType, n.Path, func(e Edge) bool {
				to, _ := g.GetNode(e.To)
				g.dfs(reverse, edgeType, &to, stack, visited)
				return true
			})
		}

	}
}

// BFS executes a depth first search with the rootNode and edge type
func (g *Graph) BFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	q := driver.NewInMemQueue()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.bfs(false, edgeType, &root, q, visited)
	q.Range(func(element interface{}) bool {
		path := element.(*Node)
		if path.Path != rootNode {
			return fn(*path)
		}
		return true
	})
}

// ReverseBFS executes a reverse depth first search with the rootNode and edge type
func (g *Graph) ReverseBFS(edgeType string, rootNode Path, fn func(nodePath Node) bool) {
	var visited = map[Path]struct{}{}
	q := driver.NewInMemQueue()
	root, ok := g.GetNode(rootNode)
	if !ok {
		return
	}
	g.bfs(true, edgeType, &root, q, visited)
	q.Range(func(element interface{}) bool {
		path := element.(*Node)
		if path.Path != rootNode {
			return fn(*path)
		}
		return true
	})
}

func (g *Graph) bfs(reverse bool, edgeType string, n *Node, q driver.Queue, visited map[Path]struct{}) {
	if _, ok := visited[n.Path]; !ok {
		visited[n.Path] = struct{}{}
		q.Enqueue(n)
		if reverse {
			g.RangeEdgesTo(edgeType, n.Path, func(e Edge) bool {
				to, _ := g.GetNode(e.From)
				g.bfs(reverse, edgeType, &to, q, visited)
				return true
			})
		} else {
			g.RangeEdgesFrom(edgeType, n.Path, func(e Edge) bool {
				to, _ := g.GetNode(e.To)
				g.bfs(reverse, edgeType, &to, q, visited)
				return true
			})
		}
	}
}

// TopologicalSort executes a topological sort
func (g *Graph) TopologicalSort(nodeType, edgeType string, fn func(node Node) bool) {
	var (
		permanent = map[Path]struct{}{}
		temp      = map[Path]struct{}{}
		stack     = driver.NewInMemStack()
	)
	g.RangeNodes(nodeType, func(n Node) bool {
		g.topology(false, edgeType, stack, n.Path, permanent, temp)
		return true
	})
	for stack.Len() > 0 {
		this, ok := stack.Pop()
		if !ok {
			return
		}
		n, _ := g.GetNode(this.(Path))
		if !fn(n) {
			return
		}
	}
}

func (g *Graph) ReverseTopologicalSort(nodeType, edgeType string, fn func(node Node) bool) {
	var (
		permanent = map[Path]struct{}{}
		temp      = map[Path]struct{}{}
		stack     = driver.NewInMemStack()
	)
	g.RangeNodes(nodeType, func(n Node) bool {
		g.topology(true, edgeType, stack, n.Path, permanent, temp)
		return true
	})
	for stack.Len() > 0 {
		this, ok := stack.Pop()
		if !ok {
			return
		}
		n, _ := g.GetNode(this.(Path))
		if !fn(n) {
			return
		}
	}
}

func (g *Graph) topology(reverse bool, edgeType string, stack driver.Stack, path Path, permanent, temporary map[Path]struct{}) {
	if _, ok := permanent[path]; ok {
		return
	}
	if reverse {
		g.RangeEdgesTo(edgeType, path, func(e Edge) bool {
			g.topology(reverse, edgeType, stack, e.From, permanent, temporary)
			return true
		})
	} else {
		g.RangeEdgesFrom(edgeType, path, func(e Edge) bool {
			g.topology(reverse, edgeType, stack, e.To, permanent, temporary)
			return true
		})
	}

	delete(temporary, path)
	permanent[path] = struct{}{}
	stack.Push(path)
}
