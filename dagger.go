package dagger

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Edge is a relationship between two nodes
type Edge[T any] struct {
	// From returns the root node of the edge
	From *Node[T] `json:"from"`
	// To returns the target node of the edge
	To *Node[T] `json:"to"`
	// Attributes returns the attributes of the edge
	Attributes map[string]string `json:"attributes"`
	id         string
	edge       *cgraph.Edge
}

// ID returns the unique identifier of the edge
func (e *Edge[T]) ID() string {
	return e.id
}

// Node is a node in the graph. It can be connected to other nodes via edges.
type Node[T any] struct {
	Value     T
	id        string
	edgesFrom *hashMap[*Edge[T]]
	edgesTo   *hashMap[*Edge[T]]
	graph     *Graph[T]
	node      *cgraph.Node
}

// ID returns the unique identifier of the node
func (n *Node[T]) ID() string {
	return n.id
}

// EdgesFrom returns the edges pointing from the current node
func (n *Node[T]) EdgesFrom(fn func(e *Edge[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
		return fn(edge)
	})
}

// EdgesTo returns the edges pointing to the current node
func (n *Node[T]) EdgesTo(fn func(e *Edge[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesTo.Range(func(key string, edge *Edge[T]) bool {
		return fn(edge)
	})
}

// SetEdge sets an edge from the current node to the node with the given nodeID
func (n *Node[T]) SetEdge(edgeID string, nodeID string, attributes map[string]string) (*Edge[T], error) {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	if attributes == nil {
		attributes = make(map[string]string)
	}
	to, ok := n.graph.GetNode(nodeID)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	if edgeID == "" {
		return nil, fmt.Errorf("edgeID cannot be empty")
	}
	e := &Edge[T]{
		id:         edgeID,
		From:       n,
		To:         to,
		Attributes: attributes,
	}
	n.graph.edges.Set(edgeID, e)
	to.edgesTo.Set(edgeID, e)
	n.edgesFrom.Set(edgeID, e)
	return e, nil
}

// AddEdge adds an edge from the current node to the node with the given nodeID
func (n *Node[T]) AddEdge(nodeID string, attributes map[string]string) (*Edge[T], error) {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	to, ok := n.graph.nodes.Get(nodeID)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}
	if attributes == nil {
		attributes = make(map[string]string)
	}
	e := &Edge[T]{
		id:         uniqueID(),
		From:       n,
		To:         to,
		Attributes: attributes,
	}
	edge, err := n.graph.viz.CreateEdge(e.id, n.node, to.node)
	if err != nil {
		return nil, err
	}
	e.edge = edge
	n.graph.edges.Set(e.id, e)
	to.edgesTo.Set(e.id, e)
	n.edgesFrom.Set(e.id, e)
	return e, nil
}

// RemoveEdge removes an edge from the current node by edgeID
func (n *Node[T]) RemoveEdge(edgeID string) {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	n.removeEdge(edgeID)
}

func (n *Node[T]) removeEdge(edgeID string) {
	edge, ok := n.graph.edges.Get(edgeID)
	if !ok {
		return
	}
	n.graph.edges.Delete(edgeID)
	n.edgesFrom.Delete(edgeID)
	n.edgesTo.Delete(edgeID)
	n.graph.viz.DeleteEdge(edge.edge)
}

// Remove removes the current node from the graph
func (n *Node[T]) Remove() error {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	n.edgesTo.Range(func(key string, edge *Edge[T]) bool {
		n.removeEdge(edge.id)
		return true
	})
	n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
		n.removeEdge(edge.id)
		return true
	})
	n.graph.nodes.Delete(n.id)
	n.graph.viz.DeleteNode(n.node)
	return nil
}

// Graph returns the graph the node belongs to
func (n *Node[T]) Graph() *Graph[T] {
	return n.graph
}

// Graph is a concurrency safe, mutable, in-memory directed graph
type Graph[T any] struct {
	nodes *hashMap[*Node[T]]
	edges *hashMap[*Edge[T]]
	gviz  *graphviz.Graphviz
	viz   *cgraph.Graph
	mu    sync.RWMutex
}

// NewGraph creates a new in-memory Graph instance
func NewGraph[T any](nodes ...*Node[T]) *Graph[T] {
	g := &Graph[T]{
		nodes: newMap[*Node[T]](),
		edges: newMap[*Edge[T]](),
		gviz:  graphviz.New(),
	}
	graph, _ := g.gviz.Graph()
	g.viz = graph
	for _, node := range nodes {
		g.SetNode(node)
	}
	for _, node := range nodes {
		node.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
			node.SetEdge(edge.id, edge.To.ID(), edge.Attributes)
			return true
		})
	}
	return g
}

// SetNode adds a node to the graph - it will generate a unique ID for the node
func (g *Graph[T]) AddNode(value T) *Node[T] {
	node := &Node[T]{
		Value:     value,
		id:        uniqueID(),
		graph:     g,
		edgesTo:   newMap[*Edge[T]](),
		edgesFrom: newMap[*Edge[T]](),
	}
	g.SetNode(node)
	return node
}

// SetNode sets a node in the graph - it will use the node's ID as the key and overwrite any existing node with the same ID
func (g *Graph[T]) SetNode(node *Node[T]) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if node.edgesFrom == nil {
		node.edgesFrom = newMap[*Edge[T]]()
	}
	if node.edgesTo == nil {
		node.edgesTo = newMap[*Edge[T]]()
	}
	node.graph = g
	if node.edgesFrom == nil {
		node.edgesFrom = newMap[*Edge[T]]()
	}
	g.nodes.Set(node.id, node)
	n, _ := g.viz.CreateNode(node.id)
	n.SetLabel(node.id)
}

// HasNode returns true if the node with the given id exists in the graph
func (g *Graph[T]) HasNode(id string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.nodes.Get(id)
	return ok
}

// HasEdge returns true if the edge with the given id exists in the graph
func (g *Graph[T]) HasEdge(id string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.edges.Get(id)
	return ok
}

// GetNode returns the node with the given id
func (g *Graph[T]) GetNode(id string) (*Node[T], bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	val, ok := g.nodes.Get(id)
	return val, ok
}

// GetEdge returns the edge with the given id
func (g *Graph[T]) GetEdge(id string) (*Edge[T], bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	val, ok := g.edges.Get(id)
	return val, ok
}

// RemoveNode removes the node with the given id from the graph
func (g *Graph[T]) RangeEdges(fn func(e *Edge[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.edges.Range(func(key string, val *Edge[T]) bool {
		return fn(val)
	})
}

// RangeNodes iterates over all nodes in the graph
func (g *Graph[T]) RangeNodes(fn func(n *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.nodes.Range(func(key string, val *Node[T]) bool {
		return fn(val)
	})
}

// DFS executes a depth first search with the rootNode and edge type
func (g *Graph[T]) DFS(rootNode string, fn func(nodestring *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var visited = map[string]struct{}{}
	stack := newStack[*Node[T]]()
	root, ok := g.nodes.Get(rootNode)
	if !ok {
		return
	}
	g.dfs(false, root, stack, visited)
	stack.Range(func(element *Node[T]) bool {
		if element.ID() != rootNode {
			return fn(element)
		}
		return true
	})
}

// ReverseDFS executes a reverse depth first search with the rootNode and edge type
func (g *Graph[T]) ReverseDFS(rootNode string, fn func(nodestring *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var visited = map[string]struct{}{}
	stack := newStack[*Node[T]]()
	root, ok := g.nodes.Get(rootNode)
	if !ok {
		return
	}
	g.dfs(true, root, stack, visited)
	stack.Range(func(element *Node[T]) bool {
		if element.ID() != rootNode {
			return fn(element)
		}
		return true
	})
}

func (g *Graph[T]) dfs(reverse bool, n *Node[T], stack *stack[*Node[T]], visited map[string]struct{}) {
	if _, ok := visited[n.ID()]; !ok {
		visited[n.ID()] = struct{}{}
		stack.Push(n)
		if reverse {
			n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
				to, _ := g.nodes.Get(edge.From.ID())
				g.dfs(reverse, to, stack, visited)
				return true
			})
		} else {
			n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
				to, _ := g.nodes.Get(edge.To.ID())
				g.dfs(reverse, to, stack, visited)
				return true
			})
		}
	}
}

// BFS executes a depth first search with the rootNode and edge type
func (g *Graph[T]) BFS(rootNode string, fn func(nodestring *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var visited = map[string]struct{}{}
	q := newQueue[*Node[T]]()
	root, ok := g.nodes.Get(rootNode)
	if !ok {
		return
	}
	g.bfs(false, root, q, visited)
	q.Range(func(element *Node[T]) bool {
		return fn(element)
	})
}

// ReverseBFS executes a reverse depth first search with the rootNode and edge type
func (g *Graph[T]) ReverseBFS(rootNode string, fn func(nodestring *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var visited = map[string]struct{}{}
	q := newQueue[*Node[T]]()
	root, ok := g.nodes.Get(rootNode)
	if !ok {
		return
	}
	g.bfs(true, root, q, visited)
	q.Range(func(element *Node[T]) bool {
		return fn(element)
	})
}

func (g *Graph[T]) bfs(reverse bool, n *Node[T], q *queue[*Node[T]], visited map[string]struct{}) {
	if _, ok := visited[n.ID()]; !ok {
		visited[n.ID()] = struct{}{}
		q.Enqueue(n)
		if reverse {
			n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
				to, _ := g.nodes.Get(edge.From.ID())
				g.bfs(reverse, to, q, visited)
				return true
			})
		} else {
			n.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
				to, _ := g.nodes.Get(edge.To.ID())
				g.bfs(reverse, to, q, visited)
				return true
			})
		}
	}
}

// TopologicalSort executes a topological sort
func (g *Graph[T]) TopologicalSort(fn func(node *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var (
		permanent = map[string]struct{}{}
		temp      = map[string]struct{}{}
		stack     = newStack[string]()
	)
	g.RangeNodes(func(n *Node[T]) bool {
		g.topology(false, stack, n, permanent, temp)
		return true
	})
	for stack.Len() > 0 {
		this, ok := stack.Pop()
		if !ok {
			return
		}
		n, _ := g.nodes.Get(this)
		if !fn(n) {
			return
		}
	}
}

// ReverseTopologicalSort executes a reverse topological sort
func (g *Graph[T]) ReverseTopologicalSort(fn func(node *Node[T]) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var (
		permanent = map[string]struct{}{}
		temp      = map[string]struct{}{}
		stack     = newStack[string]()
	)
	g.RangeNodes(func(n *Node[T]) bool {
		g.topology(true, stack, n, permanent, temp)
		return true
	})
	for stack.Len() > 0 {
		this, ok := stack.Pop()
		if !ok {
			return
		}
		n, _ := g.nodes.Get(this)
		if !fn(n) {
			return
		}
	}
}

func (g *Graph[T]) topology(reverse bool, stack *stack[string], node *Node[T], permanent, temporary map[string]struct{}) {
	if _, ok := permanent[node.id]; ok {
		return
	}
	if reverse {
		node.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	} else {
		node.edgesFrom.Range(func(key string, edge *Edge[T]) bool {
			g.topology(reverse, stack, edge.To, permanent, temporary)
			return true
		})
	}

	delete(temporary, node.id)
	permanent[node.id] = struct{}{}
	stack.Push(node.id)
}

// Vizualize returns a graphviz image
func (g *Graph[T]) Vizualize() (image.Image, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	img, err := g.gviz.RenderImage(g.viz)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// newMap creates a new generic hash map
func newMap[T any]() *hashMap[T] {
	return &hashMap[T]{
		hashMap: map[string]T{},
		mu:      sync.RWMutex{},
	}
}

// hashMap is a thread safe map
type hashMap[T any] struct {
	hashMap map[string]T
	mu      sync.RWMutex
}

// Len returns the length of the map
func (n *hashMap[T]) Len() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.hashMap)
}

// Get gets the value from the key
func (n *hashMap[T]) Get(key string) (T, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	c, ok := n.hashMap[key]
	return c, ok
}

// Set sets the key to the value
func (n *hashMap[T]) Set(key string, value T) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.hashMap[key] = value
}

// Range ranges over the map with a function until false is returned
func (n *hashMap[T]) Range(f func(key string, value T) bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for k, v := range n.hashMap {
		f(k, v)
	}
}

// Delete deletes the key from the map
func (n *hashMap[T]) Delete(key string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	delete(n.hashMap, key)
}

// Exists returns true if the key exists in the map
func (n *hashMap[T]) Exists(key string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.hashMap[key]
	return ok
}

// Clear clears the map
func (n *hashMap[T]) Clear() {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.hashMap = map[string]T{}
}

// newQueue returns a new queue with the given initial size.
func newQueue[T any]() *queue[T] {
	vals := &queue[T]{values: []T{}}
	return vals
}

// queue is a basic FIFO queue based on a circular list that resizes as needed.
type queue[T any] struct {
	mu     sync.RWMutex
	values []T
}

// Range executes a provided function once for each queue element until it returns false.
func (q *queue[T]) Range(fn func(element T) bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	for {
		r, ok := q.Dequeue()
		if !ok {
			return
		}
		if !fn(r) {
			return
		}
	}
}

// Enqueue adds an element to the end of the queue.
func (q *queue[T]) Enqueue(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.values = append(q.values, val)
}

// Dequeue removes and returns an element from the beginning of the queue.
func (q *queue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Len() == 0 {
		return *new(T), false
	}
	val := q.values[0]
	q.values = q.values[1:]
	return val, true
}

// Len returns the number of elements in the queue.
func (q *queue[T]) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.values)
}

// Peek returns the first element of the queue without removing it.
func (q *queue[T]) Peek() (T, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.Len() == 0 {
		return *new(T), false
	}
	return q.values[0], true
}

// newStack returns a new stack
func newStack[T any]() *stack[T] {
	vals := &stack[T]{values: []T{}}
	return vals
}

// stack is a basic LIFO stack
type stack[T any] struct {
	mu     sync.RWMutex
	values []T
}

// IsEmpty: check if stack is empty
func (s *stack[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s == nil || len(s.values) == 0
}

// Push a new value onto the stack
func (s *stack[T]) Push(f T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, f) // Simply append the new value to the end of the stack
}

// Remove and return top element of stack. Return false if stack is empty.
func (s *stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.IsEmpty() {
		return *new(T), false
	} else {
		index := len(s.values) - 1  // Get the index of the top most element.
		element := s.values[index]  // Index into the slice and obtain the element.
		s.values = s.values[:index] // Remove it from the stack by slicing it off.
		return element, true
	}
}

// Range executes a provided function once for each stack element until it returns false.
func (s *stack[T]) Range(fn func(element T) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for {
		r, ok := s.Pop()
		if !ok {
			return
		}
		if !fn(r) {
			return
		}
	}
}

// Len returns the number of elements in the stack.
func (s *stack[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

// Peek returns the top element of the stack without removing it. Return false if stack is empty.
func (s *stack[T]) Peek() (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.IsEmpty() {
		return *new(T), false
	} else {
		index := len(s.values) - 1 // Get the index of the top most element.
		element := s.values[index] // Index into the slice and obtain the element.
		return element, true
	}
}

// uniqueID returns a unique ID string
func uniqueID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}
