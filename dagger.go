/*
Package dagger is a collection of generic, concurrency safe datastructures including a Directed Acyclic Graph and others.
Datastructures are implemented using generics in Go 1.18.

Supported Datastructures:

DAG: thread safe directed acyclic graph

Queue: unbounded thread safe fifo queue

Stack: unbounded thread safe lifo stack

BoundedQueue: bounded thread safe fifo queue with a fixed capacity

PriorityQueue: thread safe priority queue

HashMap: thread safe hashmap

Set: thread safe set

ChannelGroup: thread safe group of channels for broadcasting 1 value to N channels

MultiContext: thread safe context for coordinating the cancellation of multiple contexts

Borrower: thread safe object ownership manager
*/
package dagger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"golang.org/x/sync/errgroup"

	"github.com/autom8ter/async"
)

// UniqueID returns a unique identifier with the given prefix
func UniqueID(prefix string) string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	if prefix == "" {
		prefix = "id"
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

// GraphEdge is a relationship between two nodes
type GraphEdge[T Node] struct {
	// ID is the unique identifier of the edge
	id string
	// Metadata is the metadata of the edge
	metadata map[string]string
	// From returns the root node of the edge
	from *GraphNode[T]
	// To returns the target node of the edge
	to *GraphNode[T]
	// Relationship is the relationship between the two nodes
	relationship string
	edge         *cgraph.Edge
}

// ID returns the unique identifier of the node
func (n *GraphEdge[T]) ID() string {
	return n.id
}

// Metadata returns the metadata of the node
func (n *GraphEdge[T]) Metadata() map[string]string {
	return n.metadata
}

// From returns the from node of the edge
func (n *GraphEdge[T]) From() *GraphNode[T] {
	return n.from
}

// To returns the to node of the edge
func (n *GraphEdge[T]) To() *GraphNode[T] {
	return n.to
}

// Relationship returns the relationship between the two nodes
func (n *GraphEdge[T]) Relationship() string {
	return n.relationship
}

// SetMetadata sets the metadata of the node
func (n *GraphEdge[T]) SetMetadata(metadata map[string]string) {
	for k, v := range metadata {
		n.metadata[k] = v
	}
}

// Node is a node in the graph. It can be connected to other nodes via edges.
type Node interface {
	// ID returns the unique identifier of the node
	ID() string
	// Metadata returns the metadata of the node
	Metadata() map[string]string
	// SetMetadata sets the metadata of the node
	SetMetadata(metadata map[string]string)
}

// GraphNode is a node in the graph. It can be connected to other nodes via edges.
type GraphNode[T Node] struct {
	Node
	edgesFrom *HashMap[string, *GraphEdge[T]]
	edgesTo   *HashMap[string, *GraphEdge[T]]
	graph     *DAG[T]
	node      *cgraph.Node
}

// DFS performs a depth-first search on the graph starting from the current node
func (n *GraphNode[T]) DFS(ctx context.Context, reverse bool, fn GraphSearchFunc[T]) error {
	return n.graph.DFS(ctx, reverse, n, fn)
}

// BFS performs a breadth-first search on the graph starting from the current node
func (n *GraphNode[T]) BFS(ctx context.Context, reverse bool, fn GraphSearchFunc[T]) error {
	return n.graph.BFS(ctx, reverse, n, fn)
}

// EdgesFrom iterates over the edges from the current node to other nodes with the given relationship.
// If the relationship is empty, all relationships will be iterated over.
func (n *GraphNode[T]) EdgesFrom(relationship string, fn func(e *GraphEdge[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if relationship != "" && edge.Relationship() != relationship {
			return true
		}
		return fn(edge)
	})
}

// EdgesTo iterates over the edges from other nodes to the current node with the given relationship.
// If the relationship is empty, all relationships will be iterated over.
func (n *GraphNode[T]) EdgesTo(relationship string, fn func(e *GraphEdge[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		if relationship != "" && edge.Relationship() != relationship {
			return true
		}
		return fn(edge)
	})
}

// SetEdge sets an edge from the current node to the node with the given nodeID.
// If the nodeID does not exist, an error is returned.
// If the edgeID is empty, a unique id will be generated.
// If the metadata is nil, an empty map will be used.
func (n *GraphNode[T]) SetEdge(relationship string, toNode Node, metadata map[string]string) (*GraphEdge[T], error) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	to, ok := n.graph.nodes.Get(toNode.ID())
	if !ok {
		to = n.graph.SetNode(toNode)
	}
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	e := &GraphEdge[T]{
		id:       strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%v-(%v)-%v", n.ID(), relationship, toNode.ID())), " ", "-"),
		metadata: metadata,
		from:     n,
		to:       to,
	}
	n.graph.edges.Set(e.ID(), e)
	to.edgesTo.Set(e.ID(), e)
	n.edgesFrom.Set(e.ID(), e)
	if n.graph.options.vizualize {
		ge, err := n.graph.viz.CreateEdge(e.ID(), n.node, to.node)
		if err != nil {
			return nil, err
		}
		ge.SetLabel(e.ID())
		if label, ok := metadata["label"]; ok {
			ge.SetLabel(label)
		}
		if color, ok := metadata["color"]; ok {
			ge.SetColor(color)
		}
		if fontColor, ok := metadata["fontcolor"]; ok {
			ge.SetFontColor(fontColor)
		}
		if weight, ok := metadata["weight"]; ok {
			weightFloat, _ := strconv.ParseFloat(weight, 64)
			ge.SetWeight(weightFloat)
		}
		if penWidth, ok := metadata["penwidth"]; ok {
			penWidthFloat, _ := strconv.ParseFloat(penWidth, 64)
			ge.SetPenWidth(penWidthFloat)
		}
		e.edge = ge
	}
	return e, nil
}

// RemoveEdge removes an edge from the current node by edgeID
func (n *GraphNode[T]) RemoveEdge(edgeID string) {
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	n.removeEdge(edgeID)
}

func (n *GraphNode[T]) removeEdge(edgeID string) {
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
func (n *GraphNode[T]) Remove() error {
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		n.removeEdge(edge.ID())
		return true
	})
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		n.removeEdge(edge.ID())
		return true
	})
	n.graph.nodes.Delete(n.ID())
	if n.graph.options.vizualize {
		n.graph.viz.DeleteNode(n.node)
	}
	return nil
}

// DirectedGraph returns the graph the node belongs to
func (n *GraphNode[T]) Graph() *DAG[T] {
	return n.graph
}

// Ancestors returns the ancestors of the current node
func (n *GraphNode[T]) Ancestors(fn func(node *GraphNode[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	visited := NewSet[string]()
	n.ancestors(visited, fn)
}

func (n *GraphNode[T]) ancestors(visited *Set[string], fn func(node *GraphNode[T]) bool) {
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited.Contains(edge.From().ID()) {
			return true
		}
		visited.Add(edge.From().ID())
		if !fn(edge.From()) {
			return false
		}
		edge.From().ancestors(visited, fn)
		return true
	})
}

// Descendants returns the descendants of the current node
func (n *GraphNode[T]) Descendants(fn func(node *GraphNode[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	visited := NewSet[string]()
	n.descendants(visited, fn)
}

func (n *GraphNode[T]) descendants(visited *Set[string], fn func(node *GraphNode[T]) bool) {
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited.Contains(edge.To().ID()) {
			return true
		}
		visited.Add(edge.To().ID())
		if !fn(edge.To()) {
			return false
		}
		edge.To().descendants(visited, fn)
		return true
	})
}

// IsConnectedTo returns true if the current node is connected to the given node in any direction
func (n *GraphNode[T]) IsConnectedTo(node *GraphNode[T]) bool {
	var result bool
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if edge.To() == node {
			result = true
			return false
		}
		return true
	})
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		if edge.From() == node {
			result = true
			return false
		}
		return true
	})
	return result
}

// DAG is a concurrency safe, mutable, in-memory directed graph
type DAG[T Node] struct {
	nodes   *HashMap[string, *GraphNode[T]]
	edges   *HashMap[string, *GraphEdge[T]]
	gviz    *graphviz.Graphviz
	viz     *cgraph.Graph
	mu      sync.RWMutex
	options *dagOpts
}

type dagOpts struct {
	vizualize bool
}

// DagOpt is an option for configuring a DAG
type DagOpt func(*dagOpts)

// WithVizualization enables graphviz visualization on the DAG
func WithVizualization() DagOpt {
	return func(opts *dagOpts) {
		opts.vizualize = true
	}
}

// NewDAG creates a new Directed Acyclic Graph instance
func NewDAG[T Node](opts ...DagOpt) (*DAG[T], error) {
	var err error
	options := &dagOpts{}
	for _, opt := range opts {
		opt(options)
	}
	g := &DAG[T]{
		nodes:   NewHashMap[string, *GraphNode[T]](),
		edges:   NewHashMap[string, *GraphEdge[T]](),
		gviz:    graphviz.New(),
		options: options,
	}
	if options.vizualize {
		graph, _ := g.gviz.Graph()
		g.viz = graph
	}
	return g, err
}

// SetNode sets a node in the graph - it will use the node's ID as the key and overwrite any existing node with the same ID
func (g *DAG[T]) SetNode(node Node) *GraphNode[T] {
	n := &GraphNode[T]{
		Node:      node,
		edgesTo:   NewHashMap[string, *GraphEdge[T]](),
		edgesFrom: NewHashMap[string, *GraphEdge[T]](),
		graph:     g,
	}
	g.nodes.Set(node.ID(), n)
	if g.options.vizualize {
		gn, err := g.viz.CreateNode(fmt.Sprintf("%v", node.ID()))
		if err != nil {
			panic(err)
		}
		gn.SetLabel(fmt.Sprintf("%v", node.ID()))
		if label, ok := node.Metadata()["label"]; ok {
			gn.SetLabel(label)
		}
		if color, ok := node.Metadata()["color"]; ok {
			gn.SetColor(color)
		}

		n.node = gn
	}

	return n
}

// HasNode returns true if the node with the given id exists in the graph
func (g *DAG[T]) HasNode(id string) bool {
	_, ok := g.nodes.Get(id)
	return ok
}

// HasEdge returns true if the edge with the given id exists in the graph
func (g *DAG[T]) HasEdge(id string) bool {
	_, ok := g.edges.Get(id)
	return ok
}

// GetNode returns the node with the given id
func (g *DAG[T]) GetNode(id string) (*GraphNode[T], bool) {
	val, ok := g.nodes.Get(id)
	return val, ok
}

// Size returns the number of nodes and edges in the graph
func (g *DAG[T]) Size() (int, int) {
	return g.nodes.Len(), g.edges.Len()
}

// GetNodes returns all nodes in the graph
func (g *DAG[T]) GetNodes() []*GraphNode[T] {
	nodes := make([]*GraphNode[T], 0, g.nodes.Len())
	g.nodes.Range(func(key string, val *GraphNode[T]) bool {
		nodes = append(nodes, val)
		return true
	})
	return nodes
}

// GetEdges returns all edges in the graph
func (g *DAG[T]) GetEdges() []*GraphEdge[T] {
	edges := make([]*GraphEdge[T], 0, g.edges.Len())
	g.edges.Range(func(key string, val *GraphEdge[T]) bool {
		edges = append(edges, val)
		return true
	})
	return edges
}

// GetEdge returns the edge with the given id
func (g *DAG[T]) GetEdge(id string) (*GraphEdge[T], bool) {
	val, ok := g.edges.Get(id)
	return val, ok
}

// RangeEdges iterates over all edges in the graph
func (g *DAG[T]) RangeEdges(fn func(e *GraphEdge[T]) bool) {
	g.edges.Range(func(key string, val *GraphEdge[T]) bool {
		return fn(val)
	})
}

// RangeNodes iterates over all nodes in the graph
func (g *DAG[T]) RangeNodes(fn func(n *GraphNode[T]) bool) {
	g.nodes.Range(func(key string, val *GraphNode[T]) bool {
		return fn(val)
	})
}

// GraphSearchFunc is a function that is called on each node in the graph during a search
type GraphSearchFunc[T Node] func(ctx context.Context, relationship string, node *GraphNode[T]) bool

// BFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *DAG[T]) BFS(ctx context.Context, reverse bool, start *GraphNode[T], search GraphSearchFunc[T]) error {
	var visited = NewSet[string]()
	stack := NewStack[*searchItem[T]]()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := g.breadthFirstSearch(ctx, &breadthFirstSearchState[T]{
		visited:      visited,
		stack:        stack,
		reverse:      reverse,
		root:         start,
		next:         start,
		relationship: "",
	}); err != nil {
		return err
	}
	stack.Range(func(element *searchItem[T]) bool {
		if element.node.ID() != start.ID() {
			return search(ctx, element.relationship, element.node)
		}
		return true
	})
	return nil
}

// DFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *DAG[T]) DFS(ctx context.Context, reverse bool, start *GraphNode[T], fn GraphSearchFunc[T]) error {
	ctx, cancel := context.WithCancel(ctx)
	var visited = NewSet[string]()
	queue := NewBoundedQueue[*searchItem[T]](0)
	egp1, ctx := errgroup.WithContext(ctx)
	egp1.Go(func() error {
		egp2, ctx := errgroup.WithContext(ctx)
		if err := g.depthFirstSearch(ctx, &depthFirstSearchState[T]{
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
		cancel()
		return nil
	})
	egp1.Go(func() error {
		queue.RangeContext(ctx, func(element *searchItem[T]) bool {
			if element.node.ID() != start.ID() {
				return fn(ctx, element.relationship, element.node)
			}
			return true
		})
		return nil
	})
	if err := egp1.Wait(); err != nil {
		return err
	}
	return nil
}

// searchItem is an item that is used in the search queue/stack in DFS/BFS
type searchItem[T Node] struct {
	node         *GraphNode[T]
	relationship string
}

type breadthFirstSearchState[T Node] struct {
	visited      *Set[string]
	stack        *Stack[*searchItem[T]]
	reverse      bool
	root         *GraphNode[T]
	next         *GraphNode[T]
	relationship string
}

func (g *DAG[T]) breadthFirstSearch(ctx context.Context, state *breadthFirstSearchState[T]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	if !state.visited.Contains(state.next.ID()) {
		state.visited.Add(state.next.ID())
		state.stack.Push(&searchItem[T]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.next.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
				g.breadthFirstSearch(ctx, &breadthFirstSearchState[T]{
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.From(),
					relationship: edge.Relationship(),
					visited:      state.visited,
				})
				return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
			})
		} else {
			state.next.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
				g.breadthFirstSearch(ctx, &breadthFirstSearchState[T]{
					visited:      state.visited,
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.To(),
					relationship: edge.Relationship(),
				})
				return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
			})
		}
	}
	return ctx.Err()
}

type depthFirstSearchState[T Node] struct {
	visited      *Set[string]
	egp          *errgroup.Group
	queue        *BoundedQueue[*searchItem[T]]
	reverse      bool
	root         *GraphNode[T]
	next         *GraphNode[T]
	relationship string
}

func (g *DAG[T]) depthFirstSearch(ctx context.Context, state *depthFirstSearchState[T]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	if !state.visited.Contains(state.next.ID()) {
		state.visited.Add(state.next.ID())
		state.queue.Push(&searchItem[T]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.egp.Go(func() error {
				state.next.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
					g.depthFirstSearch(ctx, &depthFirstSearchState[T]{
						visited:      state.visited,
						egp:          state.egp,
						queue:        state.queue,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.From(),
						relationship: edge.Relationship(),
					})
					return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
				})
				return nil
			})
		} else {
			state.egp.Go(func() error {
				state.next.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
					g.depthFirstSearch(ctx, &depthFirstSearchState[T]{
						visited:      state.visited,
						egp:          state.egp,
						queue:        state.queue,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.To(),
						relationship: edge.Relationship(),
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
func (g *DAG[T]) Acyclic() bool {
	isAcyclic := true
	g.nodes.Range(func(key string, node *GraphNode[T]) bool {
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
func (g *DAG[T]) isCyclic(node *GraphNode[T], visited *Set[string], onStack *Set[string]) bool {
	visited.Add(node.ID())
	onStack.Add(node.ID())
	result := false
	node.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited.Contains(edge.To().ID()) {
			if g.isCyclic(edge.To(), visited, onStack) {
				result = true
				return false
			}
		} else if onStack.Contains(edge.To().ID()) {
			result = true
			return false
		}
		return true
	})
	return result
}

func (g *DAG[T]) TopologicalSort(reverse bool) ([]*GraphNode[T], error) {
	if !g.Acyclic() {
		return nil, fmt.Errorf("topological sort cannot be computed on cyclical graph")
	}
	stack := NewStack[*GraphNode[T]]()
	permanent := NewSet[string]()
	temporary := NewSet[string]()
	g.nodes.Range(func(key string, node *GraphNode[T]) bool {
		g.topology(true, stack, node, permanent, temporary)
		return true
	})
	var sorted []*GraphNode[T]
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

func (g *DAG[T]) topology(reverse bool, stack *Stack[*GraphNode[T]], node *GraphNode[T], permanent, temporary *Set[string]) {
	if permanent.Contains(node.ID()) {
		return
	}
	if temporary.Contains(node.ID()) {
		panic("not a DAG")
	}
	temporary.Add(node.ID())
	if reverse {
		node.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
			g.topology(reverse, stack, edge.From(), permanent, temporary)
			return true
		})
	} else {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
			g.topology(reverse, stack, edge.From(), permanent, temporary)
			return true
		})
	}
	temporary.Remove(node.ID())
	permanent.Add(node.ID())
	stack.Push(node)
}

// GraphViz returns a graphviz image
func (g *DAG[T]) GraphViz() (image.Image, error) {
	if g.viz == nil {
		return nil, fmt.Errorf("graphviz not configured")
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	img, err := g.gviz.RenderImage(g.viz)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// NewHashMap creates a new generic hash map
func NewHashMap[K comparable, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{
		data: sync.Map{},
	}
}

// HashMap is a thread safe map
type HashMap[K comparable, V any] struct {
	data sync.Map
}

// Len returns the length of the map
func (n *HashMap[K, V]) Len() int {
	count := 0
	n.data.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Get gets the value from the key
func (n *HashMap[K, V]) Get(key K) (V, bool) {
	c, ok := n.data.Load(key)
	if !ok {
		return *new(V), ok
	}
	return c.(V), ok
}

// Set sets the key to the value
func (n *HashMap[K, V]) Set(key K, value V) {
	n.data.Store(key, value)
}

// Delete deletes the key from the map
func (n *HashMap[K, V]) Delete(key K) {
	n.data.Delete(key)
}

// Exists returns true if the key exists in the map
func (n *HashMap[K, V]) Exists(key K) bool {
	_, ok := n.Get(key)
	return ok
}

// Clear clears the map
func (n *HashMap[K, V]) Clear() {
	n.data.Range(func(key, value interface{}) bool {
		n.data.Delete(key)
		return true
	})
}

// Keys returns a copy of the keys in the map as a slice
func (n *HashMap[K, V]) Keys() []K {
	var keys []K
	n.data.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(K))
		return true
	})
	return keys
}

// Values returns a copy of the values in the map as a slice
func (n *HashMap[K, V]) Values() []V {
	var values []V
	n.data.Range(func(key, value interface{}) bool {
		values = append(values, value.(V))
		return true
	})
	return values
}

// Range ranges over the map with a function until false is returned
func (n *HashMap[K, V]) Range(f func(key K, value V) bool) {
	n.data.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

// Filter returns a new hashmap with the values that return true from the function
func (n *HashMap[K, V]) Filter(f func(key K, value V) bool) *HashMap[K, V] {
	filtered := NewHashMap[K, V]()
	n.data.Range(func(key, value interface{}) bool {
		if f(key.(K), value.(V)) {
			filtered.Set(key.(K), value.(V))
		}
		return true
	})
	return filtered
}

// Map returns a copy of the hashmap as a map[string]T
func (n *HashMap[K, V]) Map() map[K]V {
	copied := map[K]V{}
	n.data.Range(func(key, value interface{}) bool {
		copied[key.(K)] = value.(V)
		return true
	})
	return copied
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
	ch        *async.Channel[T]
}

// NewBoundedQueue returns a new BoundedQueue with the given max size. When the max size is reached, the queue will block until a value is removed.
// If maxSize is 0, the queue will always block until a value is removed. The BoundedQueue is concurrent-safe.
func NewBoundedQueue[T any](maxSize int) *BoundedQueue[T] {
	return &BoundedQueue[T]{ch: async.NewChannel[T](context.Background(), async.WithBufferSize[T](maxSize))}
}

// Range executes a provided function once for each BoundedQueue element until it returns false.
func (q *BoundedQueue[T]) Range(fn func(element T) bool) {
	for {
		value, ok := q.ch.Recv(context.Background())
		if !ok {
			return
		}
		if !fn(value) {
			return
		}
	}
}

// RangeContext executes a provided function once for each BoundedQueue element until it returns false or a value is sent to the done channel.
// Use this function when you want to continuously process items from the queue until a done signal is received.
func (q *BoundedQueue[T]) RangeContext(ctx context.Context, fn func(element T) bool) {
	for {
		value, ok := q.ch.Recv(ctx)
		if !ok {
			return
		}
		if !fn(value) {
			return
		}
	}
}

// Close closes the BoundedQueue channel.
func (q *BoundedQueue[T]) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	q.ch.Close(ctx)
}

// Push adds an element to the end of the BoundedQueue and returns a channel that will block until the element is added.
// If the queue is full, it will block until an element is removed.
func (q *BoundedQueue[T]) Push(val T) bool {
	return q.ch.Send(context.Background(), val)
}

// PushContext adds an element to the end of the BoundedQueue and returns a channel that will block until the element is added.
// If the queue is full, it will block until an element is removed or the context is cancelled.
func (q *BoundedQueue[T]) PushContext(ctx context.Context, val T) bool {
	return q.ch.Send(ctx, val)
}

// Pop removes and returns an element from the beginning of the BoundedQueue.
func (q *BoundedQueue[T]) Pop() (T, bool) {
	return q.ch.Recv(context.Background())
}

// PopContext removes and returns an element from the beginning of the BoundedQueue.
// If no element is available, it will block until an element is available or the context is cancelled.
func (q *BoundedQueue[T]) PopContext(ctx context.Context) (T, bool) {
	return q.ch.Recv(ctx)
}

// Len returns the number of elements in the BoundedQueue.
func (q *BoundedQueue[T]) Len() int {
	return q.ch.Len()
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
