package dagger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"golang.org/x/sync/errgroup"
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
	graph     *DirectedGraph[N]
	node      *cgraph.Node
}

// DFS performs a depth-first search on the graph starting from the current node
func (n *GraphNode[N]) DFS(ctx context.Context, reverse bool, fn GraphSearchFunc[N]) error {
	return n.graph.DFS(ctx, reverse, n, fn)
}

// BFS performs a breadth-first search on the graph starting from the current node
func (n *GraphNode[N]) BFS(ctx context.Context, reverse bool, fn GraphSearchFunc[N]) error {
	return n.graph.BFS(ctx, reverse, n, fn)
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

// DirectedGraph returns the graph the node belongs to
func (n *GraphNode[N]) Graph() *DirectedGraph[N] {
	return n.graph
}

// Ancestors returns the ancestors of the current node
func (n *GraphNode[N]) Ancestors(fn func(node *GraphNode[N]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	visited := make(map[string]bool)
	n.ancestors(visited, fn)
}

func (n *GraphNode[N]) ancestors(visited map[string]bool, fn func(node *GraphNode[N]) bool) {
	n.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
		if visited[edge.From.Value.ID()] {
			return true
		}
		visited[edge.From.Value.ID()] = true
		if !fn(edge.From) {
			return false
		}
		edge.From.ancestors(visited, fn)
		return true
	})
}

// Descendants returns the descendants of the current node
func (n *GraphNode[N]) Descendants(fn func(node *GraphNode[N]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	visited := make(map[string]bool)
	n.descendants(visited, fn)
}

func (n *GraphNode[N]) descendants(visited map[string]bool, fn func(node *GraphNode[N]) bool) {
	n.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
		if visited[edge.To.Value.ID()] {
			return true
		}
		visited[edge.To.Value.ID()] = true
		if !fn(edge.To) {
			return false
		}
		edge.To.descendants(visited, fn)
		return true
	})
}

// String returns a string representation of the node
func (n *GraphNode[N]) String() string {
	return fmt.Sprintf("GraphNode[%T]:%s", n.Value, n.Value.ID())
}

func (n *GraphNode[N]) IsConnectedTo(node *GraphNode[N]) bool {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	return n.isConnectedTo(node)
}

func (n *GraphNode[N]) isConnectedTo(node *GraphNode[N]) bool {
	visited := make(map[string]bool)
	return n.isConnectedToRecursive(node, visited)
}

func (n *GraphNode[N]) isConnectedToRecursive(node *GraphNode[N], visited map[string]bool) bool {
	if n == node {
		return true
	}
	visited[n.Value.ID()] = true
	for _, edge := range n.edgesFrom.Values() {
		if visited[edge.To.Value.ID()] {
			continue
		}
		if edge.To.isConnectedToRecursive(node, visited) {
			return true
		}
	}
	return false
}

// DirectedGraph is a concurrency safe, mutable, in-memory directed graph
type DirectedGraph[N Node] struct {
	nodes *HashMap[*GraphNode[N]]
	edges *HashMap[*GraphEdge[N]]
	gviz  *graphviz.Graphviz
	viz   *cgraph.Graph
	mu    sync.RWMutex
}

// NewGraph creates a new in-memory DirectedGraph instance
func NewGraph[N Node](nodes ...*GraphNode[N]) *DirectedGraph[N] {
	g := &DirectedGraph[N]{
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
func (g *DirectedGraph[N]) SetNode(node N) *GraphNode[N] {
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
func (g *DirectedGraph[N]) HasNode(id string) bool {
	_, ok := g.nodes.Get(id)
	return ok
}

// HasEdge returns true if the edge with the given id exists in the graph
func (g *DirectedGraph[N]) HasEdge(id string) bool {
	_, ok := g.edges.Get(id)
	return ok
}

// GetNode returns the node with the given id
func (g *DirectedGraph[N]) GetNode(id string) (*GraphNode[N], bool) {
	val, ok := g.nodes.Get(id)
	return val, ok
}

// Size returns the number of nodes and edges in the graph
func (g *DirectedGraph[N]) Size() (int, int) {
	return g.nodes.Len(), g.edges.Len()
}

// GetNodes returns all nodes in the graph
func (g *DirectedGraph[N]) GetNodes() []*GraphNode[N] {
	nodes := make([]*GraphNode[N], 0, g.nodes.Len())
	g.nodes.Range(func(key string, val *GraphNode[N]) bool {
		nodes = append(nodes, val)
		return true
	})
	return nodes
}

// GetEdges returns all edges in the graph
func (g *DirectedGraph[N]) GetEdges() []*GraphEdge[N] {
	edges := make([]*GraphEdge[N], 0, g.edges.Len())
	g.edges.Range(func(key string, val *GraphEdge[N]) bool {
		edges = append(edges, val)
		return true
	})
	return edges
}

// GetEdge returns the edge with the given id
func (g *DirectedGraph[N]) GetEdge(id string) (*GraphEdge[N], bool) {
	val, ok := g.edges.Get(id)
	return val, ok
}

// RemoveNode removes the node with the given id from the graph
func (g *DirectedGraph[N]) RangeEdges(fn func(e *GraphEdge[N]) bool) {
	g.edges.Range(func(key string, val *GraphEdge[N]) bool {
		return fn(val)
	})
}

// RangeNodes iterates over all nodes in the graph
func (g *DirectedGraph[N]) RangeNodes(fn func(n *GraphNode[N]) bool) {
	g.nodes.Range(func(key string, val *GraphNode[N]) bool {
		return fn(val)
	})
}

// GraphSearchFunc is a function that is called on each node in the graph during a search
type GraphSearchFunc[N Node] func(ctx context.Context, relationship string, node *GraphNode[N]) bool

// BFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *DirectedGraph[N]) BFS(ctx context.Context, reverse bool, start *GraphNode[N], search GraphSearchFunc[N]) error {
	var visited = NewSet[string]()
	stack := NewStack[*searchItem[N]]()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := g.breadthFirstSearch(ctx, &breadthFirstSearchState[N]{
		visited:      visited,
		stack:        stack,
		reverse:      reverse,
		root:         start,
		next:         start,
		relationship: "",
	}); err != nil {
		return err
	}
	stack.Range(func(element *searchItem[N]) bool {
		if element.node.Value.ID() != start.Value.ID() {
			return search(ctx, element.relationship, element.node)
		}
		return true
	})
	return nil
}

// DFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *DirectedGraph[N]) DFS(ctx context.Context, reverse bool, start *GraphNode[N], fn GraphSearchFunc[N]) error {
	var visited = NewSet[string]()
	queue := NewBoundedQueue[*searchItem[N]](0)
	var done = make(chan struct{}, 1)
	egp1, ctx := errgroup.WithContext(ctx)
	egp1.Go(func() error {
		egp2, ctx := errgroup.WithContext(ctx)
		if err := g.depthFirstSearch(ctx, &depthFirstSearchState[N]{
			visited:      visited,
			egp:          egp2,
			queue:        queue,
			reverse:      reverse,
			root:         start,
			next:         start,
			relationship: "",
		}); err != nil {
			return err
		}
		if err := egp2.Wait(); err != nil {
			return err
		}
		done <- struct{}{}
		return nil
	})
	egp1.Go(func() error {
		queue.RangeUntil(func(element *searchItem[N]) bool {
			if element.node.Value.ID() != start.Value.ID() {
				return fn(ctx, element.relationship, element.node)
			}
			return true
		}, done)
		return nil
	})
	if err := egp1.Wait(); err != nil {
		return err
	}
	return nil
}

// searchItem is an item that is used in the search queue/stack in DFS/BFS
type searchItem[N Node] struct {
	node         *GraphNode[N]
	relationship string
}

type breadthFirstSearchState[N Node] struct {
	visited      *Set[string]
	stack        *Stack[*searchItem[N]]
	reverse      bool
	root         *GraphNode[N]
	next         *GraphNode[N]
	relationship string
}

func (g *DirectedGraph[N]) breadthFirstSearch(ctx context.Context, state *breadthFirstSearchState[N]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	if !state.visited.Contains(state.next.Value.ID()) {
		state.visited.Add(state.next.Value.ID())
		state.stack.Push(&searchItem[N]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.next.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
				g.breadthFirstSearch(ctx, &breadthFirstSearchState[N]{
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.From,
					relationship: edge.Relationship,
					visited:      state.visited,
				})
				return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
			})
		} else {
			state.next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				g.breadthFirstSearch(ctx, &breadthFirstSearchState[N]{
					visited:      state.visited,
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.To,
					relationship: edge.Relationship,
				})
				return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
			})
		}
	}
	return ctx.Err()
}

type depthFirstSearchState[N Node] struct {
	visited      *Set[string]
	egp          *errgroup.Group
	queue        *BoundedQueue[*searchItem[N]]
	reverse      bool
	root         *GraphNode[N]
	next         *GraphNode[N]
	relationship string
}

func (g *DirectedGraph[N]) depthFirstSearch(ctx context.Context, state *depthFirstSearchState[N]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	if !state.visited.Contains(state.next.Value.ID()) {
		state.visited.Add(state.next.Value.ID())
		state.queue.Push(&searchItem[N]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.egp.Go(func() error {
				state.next.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
					g.depthFirstSearch(ctx, &depthFirstSearchState[N]{
						visited:      state.visited,
						egp:          state.egp,
						queue:        state.queue,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.From,
						relationship: edge.Relationship,
					})
					return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
				})
				return nil
			})
		} else {
			state.egp.Go(func() error {
				state.next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
					g.depthFirstSearch(ctx, &depthFirstSearchState[N]{
						visited:      state.visited,
						egp:          state.egp,
						queue:        state.queue,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.To,
						relationship: edge.Relationship,
					})
					return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
				})
				return nil
			})
		}
	}
	return ctx.Err()
}

// Acyclic returns true if the graph contains no cycles.
func (g *DirectedGraph[N]) Acyclic() bool {
	isAcyclic := true
	g.nodes.Range(func(key string, node *GraphNode[N]) bool {
		if node.edgesFrom.Len() > 0 {
			visited := NewSet[string]()
			onStack := NewSet[string]()
			if g.isCyclic(node, visited, onStack) {
				isAcyclic = false
				return false
			}
		}
		return true
	})
	return isAcyclic
}

// isAcyclic returns true if the graph contains no cycles.
func (g *DirectedGraph[N]) isCyclic(node *GraphNode[N], visited *Set[string], onStack *Set[string]) bool {
	visited.Add(node.Value.ID())
	onStack.Add(node.Value.ID())
	result := false
	node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
		if visited.Contains(edge.To.Value.ID()) {
			if g.isCyclic(edge.To, visited, onStack) {
				result = true
				return false
			}
		} else if onStack.Contains(edge.To.Value.ID()) {
			result = true
			return false
		}
		return true
	})
	return result
}

func (g *DirectedGraph[N]) TopologicalSort(reverse bool) ([]*GraphNode[N], error) {
	if !g.Acyclic() {
		return nil, fmt.Errorf("topological sort cannot be computed on cyclical graph")
	}
	stack := NewStack[*GraphNode[N]]()
	permanent := NewSet[string]()
	temporary := NewSet[string]()
	g.nodes.Range(func(key string, node *GraphNode[N]) bool {
		g.topology(true, stack, node, permanent, temporary)
		return true
	})
	var sorted []*GraphNode[N]
	for stack.Len() > 0 {
		val, _ := stack.Pop()
		sorted = append(sorted, val)
	}
	if reverse {
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	}
	return sorted, nil
}

func (g *DirectedGraph[N]) topology(reverse bool, stack *Stack[*GraphNode[N]], node *GraphNode[N], permanent, temporary *Set[string]) {
	if permanent.Contains(node.Value.ID()) {
		return
	}
	if temporary.Contains(node.Value.ID()) {
		panic("not a DAG")
	}
	temporary.Add(node.Value.ID())
	if reverse {
		node.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	} else {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	}
	temporary.Remove(node.Value.ID())
	permanent.Add(node.Value.ID())
	stack.Push(node)
}

// GraphViz returns a graphviz image
func (g *DirectedGraph[N]) GraphViz() (image.Image, error) {
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
func (n *HashMap[T]) Range(f func(id string, node T) bool) {
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

// Keys returns a copy of the keys in the map as a slice
func (n *HashMap[T]) Keys() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	keys := make([]string, len(n.HashMap))
	i := 0
	for k := range n.HashMap {
		keys[i] = k
		i++
	}
	return keys
}

// Values returns a copy of the values in the map as a slice
func (n *HashMap[T]) Values() []T {
	n.mu.RLock()
	defer n.mu.RUnlock()
	values := make([]T, len(n.HashMap))
	i := 0
	for _, v := range n.HashMap {
		values[i] = v
		i++
	}
	return values
}

// priorityQueueItem is an item in the priority queue
type priorityQueueItem[T any] struct {
	value    T
	priority float64
}

// PriorityQueue is a thread safe priority queue
type PriorityQueue[T any] struct {
	items []*priorityQueueItem[T]
	mu    sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: []*priorityQueueItem[T]{},
		mu:    sync.RWMutex{},
	}
}

func (q *PriorityQueue[T]) UpdatePriority(value T, priority float64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, item := range q.items {
		if reflect.DeepEqual(item.value, value) {
			q.items[i].priority = priority
			sort.Slice(q.items, func(i, j int) bool {
				return q.items[i].priority < q.items[j].priority
			})
			return
		}
	}
}

// Len returns the length of the queue
func (q *PriorityQueue[T]) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Push pushes an item onto the queue
func (q *PriorityQueue[T]) Push(item T, weight float64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, &priorityQueueItem[T]{value: item, priority: weight})
	sort.Slice(q.items, func(i, j int) bool {
		return q.items[i].priority < q.items[j].priority
	})
}

// Pop pops an item off the queue
func (q *PriorityQueue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return *new(T), false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item.value, true
}

// Peek returns the next item in the queue without removing it
func (q *PriorityQueue[T]) Peek() (T, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.items) == 0 {
		return *new(T), false
	}
	return q.items[0].value, true
}

// BoundedQueue is a basic FIFO BoundedQueue based on a buffered channel
type BoundedQueue[T any] struct {
	closeOnce sync.Once
	ch        chan T
}

// NewBoundedQueue returns a new BoundedQueue with the given max size. When the max size is reached, the queue will block until a value is removed.
// If maxSize is 0, the queue will always block until a value is removed. The BoundedQueue is concurrent-safe.
func NewBoundedQueue[T any](maxSize int) *BoundedQueue[T] {
	vals := make(chan T, maxSize)
	return &BoundedQueue[T]{ch: vals}
}

// Range executes a provided function once for each BoundedQueue element until it returns false.
func (q *BoundedQueue[T]) Range(fn func(element T) bool) {
	for {
		select {
		case r, ok := <-q.ch:
			if !ok {
				return
			}
			if !fn(r) {
				return
			}
		default:
			if len(q.ch) == 0 {
				return
			}
		}
	}
}

// RangeUntil executes a provided function once for each BoundedQueue element until it returns false or a value is sent to the done channel.
// Use this function when you want to continuously process items from the queue until a done signal is received.
func (q *BoundedQueue[T]) RangeUntil(fn func(element T) bool, done chan struct{}) {
	for {
		select {
		case <-done:
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

// Close closes the BoundedQueue channel.
func (q *BoundedQueue[T]) Close() {
	q.closeOnce.Do(func() {
		close(q.ch)
	})
}

// Push adds an element to the end of the BoundedQueue.
func (q *BoundedQueue[T]) Push(val T) {
	q.ch <- val
}

// Pop removes and returns an element from the beginning of the BoundedQueue.
func (q *BoundedQueue[T]) Pop() (T, bool) {
	select {
	case r, ok := <-q.ch:
		return r, ok
	}
}

// Len returns the number of elements in the BoundedQueue.
func (q *BoundedQueue[T]) Len() int {
	return len(q.ch)
}

// Queue is a thread safe non-blocking queue
type Queue[T any] struct {
	mu     sync.RWMutex
	values []T
}

// NewQueue returns a new Queue
func NewQueue[T any]() *Queue[T] {
	vals := &Queue[T]{values: []T{}}
	return vals
}

// Push a new value onto the Queue
func (s *Queue[T]) Push(f T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, f) // Simply append the new value to the end of the Queue
}

// Pop and return top element of Queue. Return false if Queue is empty.
func (s *Queue[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.values) == 0 {
		return *new(T), false
	} else {
		index := len(s.values) - 1
		val := s.values[index]
		s.values = s.values[:index]
		return val, true
	}
}

// Len returns the length of the queue
func (s *Queue[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

// Peek returns the next item in the queue without removing it
func (s *Queue[T]) Peek() (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.values) == 0 {
		return *new(T), false
	}
	return s.values[len(s.values)-1], true
}

// Range executes a provided function once for each Queue element until it returns false or the Queue is empty.
func (q *Queue[T]) Range(fn func(element T) bool) {
	for {
		val, ok := q.Pop()
		if !ok {
			return
		}
		if !fn(val) {
			return
		}
	}
}

// RangeUntil executes a provided function once for each Queue element until it returns false or a value is sent on the done channel.
// Use this function when you want to continuously process items from the queue until a done signal is received.
func (q *Queue[T]) RangeUntil(fn func(element T) bool, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			val, ok := q.Pop()
			if !ok {
				return
			}
			if !fn(val) {
				return
			}
		}
	}
}

// NewStack returns a new Stack instance
func NewStack[T any]() *Stack[T] {
	vals := &Stack[T]{values: []T{}}
	return vals
}

// Stack is a basic LIFO Stack
type Stack[T any] struct {
	mu     sync.RWMutex
	values []T
}

// RangeUntil executes a provided function once after calling Pop on the stack until the function returns false or a value is sent on the done channel.
// Use this function when you want to continuously process items from the stack until a done signal is received.
func (s *Stack[T]) RangeUntil(fn func(element T) bool, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			val, ok := s.Pop()
			if !ok {
				return
			}
			if !fn(val) {
				return
			}
		}
	}
}

// Push a new value onto the Stack (LIFO)
func (s *Stack[T]) Push(f T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, f) // Simply append the new value to the end of the Stack
}

// Pop removes and return top element of Stack. Return false if Stack is empty.
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

// Clear removes all elements from the Stack
func (s *Stack[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = []T{}
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

// Values returns the values of the stack as an array
func (s *Stack[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values
}

// Sort returns the values of the stack as an array sorted by the provided less function
func (s *Stack[T]) Sort(lessFunc func(i T, j T) bool) []T {
	values := s.Values()
	sort.Slice(values, func(i, j int) bool {
		return lessFunc(values[i], values[j])
	})
	return values
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

// Set is a basic thread-safe Set implementation.
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

// Values returns the values of the set as an array
func (s *Set[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]T, len(s.values))
	for k, _ := range s.values {
		values = append(values, k)
	}
	return values
}

// Sort returns the values of the set as an array sorted by the provided less function
func (s *Set[T]) Sort(lessFunc func(i T, j T) bool) []T {
	values := s.Values()
	sort.Slice(values, func(i, j int) bool {
		return lessFunc(values[i], values[j])
	})
	return values
}

// debugF logs a message if the DAGGER_DEBUG environment variable is set.
// It adds a stacktrace to the log message.
func debugF(format string, a ...interface{}) {
	if os.Getenv("DAGGER_DEBUG") != "" {
		format = fmt.Sprintf("DEBUG: %s\n", format)
		log.Printf(format, a...)
	}
}
