package dagger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Node is a node in a graph
type Node interface {
	// ID returns the unique id of the node
	ID() string
}

// String is a generic string with methods to satisfy various interfaces such as the Node interface
type String string

// UniqueID returns a unique identifier with the given prefix
func UniqueID(prefix string) String {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	if prefix == "" {
		prefix = "id"
	}
	return String(fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b)))
}

// ID returns the string
func (s String) ID() string {
	return string(s)
}

// GraphEdge is a relationship between two nodes
type GraphEdge[N Node] struct {
	// ID is the unique identifier of the edge
	ID string `json:"id"`
	// Metadata is the metadata of the edge
	Metadata map[string]string `json:"metadata"`
	// From returns the root node of the edge
	From *GraphNode[N] `json:"from"`
	// To returns the target node of the edge
	To *GraphNode[N] `json:"to"`
	// Relationship is the relationship between the two nodes
	Relationship string `json:"relationship"`
	edge         *cgraph.Edge
}

// GraphNode is a node in the graph. It can be connected to other nodes via edges.
type GraphNode[N Node] struct {
	// Value is the value of the node
	Value     N
	edgesFrom *HashMap[*GraphEdge[N]]
	edgesTo   *HashMap[*GraphEdge[N]]
	graph     *Graph[N]
	node      *cgraph.Node
}

// EdgesFrom returns the edges pointing from the current node
func (n *GraphNode[N]) EdgesFrom(fn func(e *GraphEdge[N]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
		return fn(edge)
	})
}

// EdgesTo returns the edges pointing to the current node
func (n *GraphNode[N]) EdgesTo(fn func(e *GraphEdge[N]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
		return fn(edge)
	})
}

// SetEdge sets an edge from the current node to the node with the given nodeID.
// If the nodeID does not exist, an error is returned.
// If the edgeID is empty, a unique id will be generated.
// If the metadata is nil, an empty map will be used.
func (n *GraphNode[N]) SetEdge(toNode *GraphNode[N], relationship string, metadata map[string]string) (*GraphEdge[N], error) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	e := &GraphEdge[N]{
		ID:       strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s-(%s)-%s", n.Value.ID(), relationship, toNode.Value.ID())), " ", "-"),
		Metadata: metadata,
		From:     n,
		To:       toNode,
	}
	n.graph.edges.Set(e.ID, e)
	toNode.edgesTo.Set(e.ID, e)
	n.edgesFrom.Set(e.ID, e)
	return e, nil
}

// RemoveEdge removes an edge from the current node by edgeID
func (n *GraphNode[N]) RemoveEdge(edgeID string) {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	n.removeEdge(edgeID)
}

func (n *GraphNode[N]) removeEdge(edgeID string) {
	edge, ok := n.graph.edges.Get(edgeID)
	if !ok {
		return
	}
	n.graph.edges.Delete(edgeID)
	n.edgesFrom.Delete(edgeID)
	n.edgesTo.Delete(edgeID)
	if edge.edge != nil {
		n.graph.viz.DeleteEdge(edge.edge)
	}
}

// Remove removes the current node from the graph
func (n *GraphNode[N]) Remove() error {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	n.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
		n.removeEdge(edge.ID)
		return true
	})
	n.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
		n.removeEdge(edge.ID)
		return true
	})
	n.graph.nodes.Delete(n.Value.ID())
	n.graph.viz.DeleteNode(n.node)
	return nil
}

// Graph returns the graph the node belongs to
func (n *GraphNode[N]) Graph() *Graph[N] {
	return n.graph
}

// Graph is a concurrency safe, mutable, in-memory directed graph
type Graph[N Node] struct {
	nodes *HashMap[*GraphNode[N]]
	edges *HashMap[*GraphEdge[N]]
	gviz  *graphviz.Graphviz
	viz   *cgraph.Graph
	mu    sync.RWMutex
}

// NewGraph creates a new in-memory Graph instance
func NewGraph[N Node](nodes ...*GraphNode[N]) *Graph[N] {
	g := &Graph[N]{
		nodes: NewHashMap[*GraphNode[N]](),
		edges: NewHashMap[*GraphEdge[N]](),
		gviz:  graphviz.New(),
	}
	graph, _ := g.gviz.Graph()
	g.viz = graph
	for _, node := range nodes {
		g.SetNode(node.Value)
	}
	for _, node := range nodes {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
			if _, err := node.SetEdge(edge.To, edge.Relationship, edge.Metadata); err != nil {
				panic(err)
			}
			return true
		})
	}
	return g
}

// SetNode sets a node in the graph - it will use the node's ID as the key and overwrite any existing node with the same ID
func (g *Graph[N]) SetNode(node N) *GraphNode[N] {
	n := &GraphNode[N]{
		Value:     node,
		edgesTo:   NewHashMap[*GraphEdge[N]](),
		edgesFrom: NewHashMap[*GraphEdge[N]](),
		graph:     g,
	}
	g.nodes.Set(node.ID(), n)
	gn, err := g.viz.CreateNode(node.ID())
	if err != nil {
		panic(err)
	}
	gn.SetLabel(node.ID())
	n.node = gn
	return n
}

// HasNode returns true if the node with the given id exists in the graph
func (g *Graph[N]) HasNode(id string) bool {
	_, ok := g.nodes.Get(id)
	return ok
}

// HasEdge returns true if the edge with the given id exists in the graph
func (g *Graph[N]) HasEdge(id string) bool {
	_, ok := g.edges.Get(id)
	return ok
}

// GetNode returns the node with the given id
func (g *Graph[N]) GetNode(id string) (*GraphNode[N], bool) {
	val, ok := g.nodes.Get(id)
	return val, ok
}

// Size returns the number of nodes and edges in the graph
func (g *Graph[N]) Size() (int, int) {
	return g.nodes.Len(), g.edges.Len()
}

// GetNodes returns all nodes in the graph
func (g *Graph[N]) GetNodes() []*GraphNode[N] {
	nodes := make([]*GraphNode[N], 0, g.nodes.Len())
	g.nodes.Range(func(key string, val *GraphNode[N]) bool {
		nodes = append(nodes, val)
		return true
	})
	return nodes
}

// GetEdges returns all edges in the graph
func (g *Graph[N]) GetEdges() []*GraphEdge[N] {
	edges := make([]*GraphEdge[N], 0, g.edges.Len())
	g.edges.Range(func(key string, val *GraphEdge[N]) bool {
		edges = append(edges, val)
		return true
	})
	return edges
}

// GetEdge returns the edge with the given id
func (g *Graph[N]) GetEdge(id string) (*GraphEdge[N], bool) {
	val, ok := g.edges.Get(id)
	return val, ok
}

// RemoveNode removes the node with the given id from the graph
func (g *Graph[N]) RangeEdges(fn func(e *GraphEdge[N]) bool) {
	g.edges.Range(func(key string, val *GraphEdge[N]) bool {
		return fn(val)
	})
}

// RangeNodes iterates over all nodes in the graph
func (g *Graph[N]) RangeNodes(fn func(n *GraphNode[N]) bool) {
	g.nodes.Range(func(key string, val *GraphNode[N]) bool {
		return fn(val)
	})
}

// BFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *Graph[N]) BFS(reverse bool, start *GraphNode[N], fn func(node *GraphNode[N]) bool) {
	var visited = map[string]struct{}{}
	q := NewQueue[*GraphNode[N]]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		var (
			wg = &sync.WaitGroup{}
			mu = &sync.RWMutex{}
		)
		g.bfs(wg, mu, reverse, start, nil, q, visited)
		wg.Wait()
		cancel()
	}()
	q.RangeContext(ctx, func(element *GraphNode[N]) bool {
		if element.Value.ID() != start.Value.ID() {
			return fn(element)
		}
		return true
	})
}

// DFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *Graph[N]) DFS(reverse bool, start *GraphNode[N], fn func(node *GraphNode[N]) bool) {
	var visited = map[string]struct{}{}
	stack := NewStack[*GraphNode[N]]()
	g.dfs(reverse, start, nil, stack, visited)
	stack.Range(func(element *GraphNode[N]) bool {
		if element.Value.ID() != start.Value.ID() {
			return fn(element)
		}
		return true
	})
}

func (g *Graph[N]) dfs(reverse bool, root, next *GraphNode[N], stack *Stack[*GraphNode[N]], visited map[string]struct{}) {
	if next == nil {
		next = root
	}
	if _, ok := visited[next.Value.ID()]; !ok {
		visited[next.Value.ID()] = struct{}{}
		stack.Push(next)
		if reverse {
			next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				to, _ := g.nodes.Get(edge.From.Value.ID())
				g.dfs(reverse, root, to, stack, visited)
				return len(visited) < g.nodes.Len()
			})
		} else {
			next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				to, _ := g.nodes.Get(edge.To.Value.ID())
				g.dfs(reverse, root, to, stack, visited)
				return len(visited) < g.nodes.Len()
			})
		}
	}
	debugF("dfs: finished visiting %s\n", next.Value.ID())
}

func (g *Graph[N]) bfs(wg *sync.WaitGroup, mu *sync.RWMutex, reverse bool, root, next *GraphNode[N], q *Queue[*GraphNode[N]], visited map[string]struct{}) {
	if next == nil {
		next = root
	}
	mu.Lock()
	_, ok := visited[next.Value.ID()]
	if !ok {
		visited[next.Value.ID()] = struct{}{}
	}
	mu.Unlock()
	if !ok {
		q.Push(next)
		if reverse {
			next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				to, _ := g.nodes.Get(edge.From.Value.ID())
				wg.Add(1)
				go func(to *GraphNode[N], visited map[string]struct{}) {
					defer wg.Done()
					g.bfs(wg, mu, reverse, root, to, q, visited)
				}(to, visited)

				return len(visited) < g.nodes.Len()
			})
		} else {
			next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				to, _ := g.nodes.Get(edge.To.Value.ID())
				wg.Add(1)
				go func(to *GraphNode[N], visited map[string]struct{}) {
					defer wg.Done()
					g.bfs(wg, mu, reverse, root, to, q, visited)
				}(to, visited)
				return len(visited) < g.nodes.Len()
			})
		}
	}
	debugF("bfs: finished visiting %s\n", next.Value.ID())
}

// TopologicalSort executes a topological sort
func (g *Graph[N]) TopologicalSort(reverse bool, fn func(node *GraphNode[N]) bool) {
	var (
		permanent = map[string]struct{}{}
		temp      = map[string]struct{}{}
		stack     = NewStack[string]()
	)
	g.RangeNodes(func(n *GraphNode[N]) bool {
		g.topology(reverse, stack, n, permanent, temp)
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

func (g *Graph[N]) topology(reverse bool, stack *Stack[string], node *GraphNode[N], permanent, temporary map[string]struct{}) {
	if _, ok := permanent[node.Value.ID()]; ok {
		return
	}
	if reverse {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	} else {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
			g.topology(reverse, stack, edge.To, permanent, temporary)
			return true
		})
	}

	delete(temporary, node.Value.ID())
	permanent[node.Value.ID()] = struct{}{}
	stack.Push(node.Value.ID())
}

// GraphViz returns a graphviz image
func (g *Graph[N]) GraphViz() (image.Image, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	img, err := g.gviz.RenderImage(g.viz)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// NewHashMap creates a new generic hash map
func NewHashMap[T any]() *HashMap[T] {
	return &HashMap[T]{
		HashMap: map[string]T{},
		mu:      sync.RWMutex{},
	}
}

// HashMap is a thread safe map
type HashMap[T any] struct {
	HashMap map[string]T
	mu      sync.RWMutex
}

// Len returns the length of the map
func (n *HashMap[T]) Len() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.HashMap)
}

// Get gets the value from the key
func (n *HashMap[T]) Get(key string) (T, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	c, ok := n.HashMap[key]
	return c, ok
}

// Set sets the key to the value
func (n *HashMap[T]) Set(key string, value T) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.HashMap[key] = value
}

// Range ranges over the map with a function until false is returned
func (n *HashMap[T]) Range(f func(key string, value T) bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for k, v := range n.HashMap {
		f(k, v)
	}
}

// Delete deletes the key from the map
func (n *HashMap[T]) Delete(key string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	delete(n.HashMap, key)
}

// Exists returns true if the key exists in the map
func (n *HashMap[T]) Exists(key string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.HashMap[key]
	return ok
}

// Clear clears the map
func (n *HashMap[T]) Clear() {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.HashMap = map[string]T{}
}

// NewQueue returns a new Queue with the given initial size.
func NewQueue[T any]() *Queue[T] {
	vals := make(chan T)
	return &Queue[T]{ch: vals}
}

// Queue is a basic FIFO Queue based on a circular list that resizes as needed.
type Queue[T any] struct {
	closeOnce sync.Once
	ch        chan T
}

// Range executes a provided function once for each Queue element until it returns false.
func (q *Queue[T]) Range(fn func(element T) bool) {
	for {
		select {
		case r, ok := <-q.ch:
			if !ok {
				return
			}
			if !fn(r) {
				return
			}
		}
	}
}

// RangeContext executes a provided function once for each Queue element until it returns false or the context is done.
func (q *Queue[T]) RangeContext(ctx context.Context, fn func(element T) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-q.ch:
			if !ok {
				return
			}
			if !fn(r) {
				return
			}
		}
	}
}

// Close closes the Queue channel.
func (q *Queue[T]) Close() {
	q.closeOnce.Do(func() {
		close(q.ch)
	})
}

// Push adds an element to the end of the Queue.
func (q *Queue[T]) Push(val T) {
	q.ch <- val
}

// Pop removes and returns an element from the beginning of the Queue.
func (q *Queue[T]) Pop() (T, bool) {
	select {
	case r, ok := <-q.ch:
		return r, ok
	}
}

// Len returns the number of elements in the Queue.
func (q *Queue[T]) Len() int {
	return len(q.ch)
}

// NewStack returns a new Stack
func NewStack[T any]() *Stack[T] {
	vals := &Stack[T]{values: []T{}}
	return vals
}

// Stack is a basic LIFO Stack
type Stack[T any] struct {
	mu     sync.RWMutex
	values []T
}

// Push a new value onto the Stack
func (s *Stack[T]) Push(f T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, f) // Simply append the new value to the end of the Stack
}

// Remove and return top element of Stack. Return false if Stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.values) == 0 {
		return *new(T), false
	} else {
		index := len(s.values) - 1  // Get the index of the top most element.
		element := s.values[index]  // Index into the slice and obtain the element.
		s.values = s.values[:index] // Remove it from the Stack by slicing it off.
		return element, true
	}
}

// Range executes a provided function once for each Stack element until it returns false.
func (s *Stack[T]) Range(fn func(element T) bool) {
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

// Len returns the number of elements in the Stack.
func (s *Stack[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

// Peek returns the top element of the Stack without removing it. Return false if Stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.values) == 0 {
		return *new(T), false
	} else {
		index := len(s.values) - 1 // Get the index of the top most element.
		element := s.values[index] // Index into the slice and obtain the element.
		return element, true
	}
}

type Set[T comparable] struct {
	mu     sync.RWMutex
	values map[T]struct{}
}

// NewSet returns a new Set with the given initial size.
func NewSet[T comparable]() *Set[T] {
	vals := &Set[T]{values: map[T]struct{}{}}
	return vals
}

// Add adds an element to the Set.
func (s *Set[T]) Add(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[val] = struct{}{}
}

// Remove removes an element from the Set.
func (s *Set[T]) Remove(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, val)
}

// Contains returns true if the Set contains the element.
func (s *Set[T]) Contains(val T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.values[val]
	return ok
}

// Range executes a provided function once for each Set element until it returns false.
func (s *Set[T]) Range(fn func(element T) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.values {
		if !fn(k) {
			return
		}
	}
}

// Len returns the number of elements in the Set.
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

// debugF logs a message if the DAGGER_DEBUG environment variable is set.
// It adds a stacktrace to the log message.
func debugF(format string, a ...interface{}) {
	if os.Getenv("DAGGER_DEBUG") != "" {
		format = fmt.Sprintf("DEBUG: %s\n", format)
		log.Printf(format, a...)
	}
}
