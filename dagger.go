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
*/
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
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"golang.org/x/sync/errgroup"
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
type GraphEdge[T any] struct {
	// ID is the unique identifier of the edge
	ID string `json:"id"`
	// Metadata is the metadata of the edge
	Metadata map[string]string `json:"metadata"`
	// From returns the root node of the edge
	From *GraphNode[T] `json:"from"`
	// To returns the target node of the edge
	To *GraphNode[T] `json:"to"`
	// Relationship is the relationship between the two nodes
	Relationship string `json:"relationship"`
	edge         *cgraph.Edge
}

// GraphNode is a node in the graph. It can be connected to other nodes via edges.
type GraphNode[T any] struct {
	id        string
	metadata  map[string]string
	value     T
	edgesFrom *HashMap[*GraphEdge[T]]
	edgesTo   *HashMap[*GraphEdge[T]]
	graph     *DAG[T]
	node      *cgraph.Node
}

// ID returns the unique identifier of the node
func (n *GraphNode[T]) ID() string {
	return n.id
}

// Metadata returns the metadata of the node
func (n *GraphNode[T]) Metadata() map[string]string {
	return n.metadata
}

// Value returns the value of the node
func (n *GraphNode[T]) Value() T {
	return n.value
}

// SetMetadata sets the metadata of the node
func (n *GraphNode[T]) SetMetadata(metadata map[string]string) {
	n.metadata = metadata
}

// SetID sets the unique identifier of the node
func (n *GraphNode[T]) SetID(id string) {
	n.id = id
}

// SetValue sets the value of the node
func (n *GraphNode[T]) SetValue(value T) {
	n.value = value
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
		if relationship != "" && edge.Relationship != relationship {
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
		if relationship != "" && edge.Relationship != relationship {
			return true
		}
		return fn(edge)
	})
}

// SetEdge sets an edge from the current node to the node with the given nodeID.
// If the nodeID does not exist, an error is returned.
// If the edgeID is empty, a unique id will be generated.
// If the metadata is nil, an empty map will be used.
func (n *GraphNode[T]) SetEdge(toNode *GraphNode[T], relationship string, metadata map[string]string) (*GraphEdge[T], error) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	n.graph.mu.Lock()
	defer n.graph.mu.Unlock()
	e := &GraphEdge[T]{
		ID:       strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s-(%s)-%s", n.id, relationship, toNode.id)), " ", "-"),
		Metadata: metadata,
		From:     n,
		To:       toNode,
	}
	n.graph.edges.Set(e.ID, e)
	toNode.edgesTo.Set(e.ID, e)
	n.edgesFrom.Set(e.ID, e)
	if n.graph.options.vizualize {
		ge, err := n.graph.viz.CreateEdge(e.ID, n.node, toNode.node)
		if err != nil {
			return nil, err
		}
		ge.SetLabel(e.ID)
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
		n.removeEdge(edge.ID)
		return true
	})
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		n.removeEdge(edge.ID)
		return true
	})
	n.graph.nodes.Delete(n.id)
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
	visited := make(map[string]bool)
	n.ancestors(visited, fn)
}

func (n *GraphNode[T]) ancestors(visited map[string]bool, fn func(node *GraphNode[T]) bool) {
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited[edge.From.id] {
			return true
		}
		visited[edge.From.id] = true
		if !fn(edge.From) {
			return false
		}
		edge.From.ancestors(visited, fn)
		return true
	})
}

// Descendants returns the descendants of the current node
func (n *GraphNode[T]) Descendants(fn func(node *GraphNode[T]) bool) {
	n.graph.mu.RLock()
	defer n.graph.mu.RUnlock()
	visited := make(map[string]bool)
	n.descendants(visited, fn)
}

func (n *GraphNode[T]) descendants(visited map[string]bool, fn func(node *GraphNode[T]) bool) {
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited[edge.To.id] {
			return true
		}
		visited[edge.To.id] = true
		if !fn(edge.To) {
			return false
		}
		edge.To.descendants(visited, fn)
		return true
	})
}

// String returns a string representation of the node
func (n *GraphNode[T]) String() string {
	return fmt.Sprintf("GraphNode[%T]:%s", n.value, n.id)
}

// IsConnectedTo returns true if the current node is connected to the given node in any direction
func (n *GraphNode[T]) IsConnectedTo(node *GraphNode[T]) bool {
	var result bool
	n.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if edge.To == node {
			result = true
			return false
		}
		return true
	})
	n.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
		if edge.From == node {
			result = true
			return false
		}
		return true
	})
	return result
}

// DAG is a concurrency safe, mutable, in-memory directed graph
type DAG[T any] struct {
	nodes   *HashMap[*GraphNode[T]]
	edges   *HashMap[*GraphEdge[T]]
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
func NewDAG[T any](opts ...DagOpt) (*DAG[T], error) {
	var err error
	options := &dagOpts{}
	for _, opt := range opts {
		opt(options)
	}
	g := &DAG[T]{
		nodes:   NewHashMap[*GraphNode[T]](),
		edges:   NewHashMap[*GraphEdge[T]](),
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
func (g *DAG[T]) SetNode(id string, node T, metadata map[string]string) *GraphNode[T] {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	n := &GraphNode[T]{
		id:        id,
		value:     node,
		metadata:  metadata,
		edgesTo:   NewHashMap[*GraphEdge[T]](),
		edgesFrom: NewHashMap[*GraphEdge[T]](),
		graph:     g,
	}
	g.nodes.Set(id, n)
	if g.options.vizualize {
		gn, err := g.viz.CreateNode(id)
		if err != nil {
			panic(err)
		}
		gn.SetLabel(id)
		if label, ok := metadata["label"]; ok {
			gn.SetLabel(label)
		}
		if color, ok := metadata["color"]; ok {
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
type GraphSearchFunc[T any] func(ctx context.Context, relationship string, node *GraphNode[T]) bool

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
		if element.node.id != start.id {
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
	var visited = NewSet[string]()
	queue := NewBoundedQueue[*searchItem[T]](0)
	var done = make(chan struct{}, 1)
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
		done <- struct{}{}
		return nil
	})
	egp1.Go(func() error {
		queue.RangeUntil(func(element *searchItem[T]) bool {
			if element.node.id != start.id {
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
type searchItem[T any] struct {
	node         *GraphNode[T]
	relationship string
}

type breadthFirstSearchState[T any] struct {
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
	if !state.visited.Contains(state.next.id) {
		state.visited.Add(state.next.id)
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
					next:         edge.From,
					relationship: edge.Relationship,
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
					next:         edge.To,
					relationship: edge.Relationship,
				})
				return state.visited.Len() < g.nodes.Len() && ctx.Err() == nil
			})
		}
	}
	return ctx.Err()
}

type depthFirstSearchState[T any] struct {
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
	if !state.visited.Contains(state.next.id) {
		state.visited.Add(state.next.id)
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
						next:         edge.From,
						relationship: edge.Relationship,
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
	visited.Add(node.id)
	onStack.Add(node.id)
	result := false
	node.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
		if visited.Contains(edge.To.id) {
			if g.isCyclic(edge.To, visited, onStack) {
				result = true
				return false
			}
		} else if onStack.Contains(edge.To.id) {
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
	if permanent.Contains(node.id) {
		return
	}
	if temporary.Contains(node.id) {
		panic("not a DAG")
	}
	temporary.Add(node.id)
	if reverse {
		node.edgesTo.Range(func(key string, edge *GraphEdge[T]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	} else {
		node.edgesFrom.Range(func(key string, edge *GraphEdge[T]) bool {
			g.topology(reverse, stack, edge.From, permanent, temporary)
			return true
		})
	}
	temporary.Remove(node.id)
	permanent.Add(node.id)
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
func NewHashMap[T any]() *HashMap[T] {
	return &HashMap[T]{
		data: map[string]T{},
		mu:   sync.RWMutex{},
	}
}

// HashMap is a thread safe map
type HashMap[T any] struct {
	data map[string]T
	mu   sync.RWMutex
}

// Len returns the length of the map
func (n *HashMap[T]) Len() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.data)
}

// Get gets the value from the key
func (n *HashMap[T]) Get(key string) (T, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	c, ok := n.data[key]
	return c, ok
}

// Set sets the key to the value
func (n *HashMap[T]) Set(key string, value T) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.data[key] = value
}

// Range ranges over the map with a function until false is returned
func (n *HashMap[T]) Range(f func(id string, node T) bool) {
	n.mu.RLock()
	data := n.data
	n.mu.RUnlock()
	for k, v := range data {
		if !f(k, v) {
			return
		}
	}
}

// Delete deletes the key from the map
func (n *HashMap[T]) Delete(key string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.data, key)
}

// Exists returns true if the key exists in the map
func (n *HashMap[T]) Exists(key string) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.data[key]
	return ok
}

// Clear clears the map
func (n *HashMap[T]) Clear() {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.data = map[string]T{}
}

// Keys returns a copy of the keys in the map as a slice
func (n *HashMap[T]) Keys() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	keys := make([]string, len(n.data))
	i := 0
	for k := range n.data {
		keys[i] = k
		i++
	}
	return keys
}

// Values returns a copy of the values in the map as a slice
func (n *HashMap[T]) Values() []T {
	n.mu.RLock()
	defer n.mu.RUnlock()
	values := make([]T, len(n.data))
	i := 0
	for _, v := range n.data {
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

// ChannelGroup is a thread-safe group of channels. It is useful for broadcasting a value to multiple channels at once.
type ChannelGroup[T any] struct {
	ctx         *MultiContext
	subscribers *HashMap[channelGroupState[T]]
	wg          sync.WaitGroup
}

type channelGroupState[T any] struct {
	ch  chan T
	ctx context.Context
}

// NewChannelGroup returns a new ChannelGroup. The context is used to cancel all subscribers when the context is canceled.
// A channel group is useful for broadcasting a value to multiple subscribers.
func NewChannelGroup[T any](ctx context.Context) *ChannelGroup[T] {
	return &ChannelGroup[T]{
		ctx:         NewMultiContext(ctx),
		subscribers: NewHashMap[channelGroupState[T]](),
		wg:          sync.WaitGroup{},
	}
}

// Send sends a value to all channels in the group.
func (b *ChannelGroup[T]) Send(ctx context.Context, val T) {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.subscribers.Range(func(key string, state channelGroupState[T]) bool {
			select {
			case <-state.ctx.Done():
				return false
			case state.ch <- val:
				return true
			}
		})
	}()
	return
}

// Channel returns a channel that will receive values from broadcasted values. The channel will be closed when the context is canceled.
// This is a non-blocking operation.
func (b *ChannelGroup[T]) Channel(ctx context.Context) <-chan T {
	ch := make(chan T, 1)
	id := UniqueID("subscriber")
	b.subscribers.Set(id, channelGroupState[T]{
		ch:  ch,
		ctx: b.ctx.WithContext(ctx),
	})
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		<-b.ctx.WithContext(ctx).Done()
		b.subscribers.Delete(id)
		close(ch)
	}()
	return ch
}

// Len returns the number of subscribers.
func (c *ChannelGroup[T]) Len() int {
	return c.subscribers.Len()
}

// Close blocks until all subscribers have been removed and then closes the broadcast.
func (b *ChannelGroup[T]) Close() {
	b.ctx.Cancel()
	b.wg.Wait()
}

// MultiContext is a context that can be used to combine contexts with a root context so they can be cancelled together.
type MultiContext struct {
	context.Context
	mu      sync.Mutex
	cancel  context.CancelFunc
	cancels []context.CancelFunc
}

// NewMultiContext returns a new MultiContext.
func NewMultiContext(ctx context.Context) *MultiContext {
	ctx, cancel := context.WithCancel(ctx)
	m := &MultiContext{
		Context: ctx,
		cancel:  cancel,
	}
	go func() {
		select {
		case <-m.Done():
			m.mu.Lock()
			for _, cancel := range m.cancels {
				cancel()
			}
			m.mu.Unlock()
		}
	}()
	return m
}

// WithContext returns a new context that is a child of the root context.
// This context will be cancelled when the multi context is cancelled.
func (m *MultiContext) WithContext(ctx context.Context) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, cancel := context.WithCancel(ctx)
	m.cancels = append(m.cancels, cancel)
	return ctx
}

// Cancel cancels all child contexts.
func (m *MultiContext) Cancel() {
	for _, cancel := range m.cancels {
		cancel()
	}
	m.cancel()
}

// debugF logs a message if the DAGGER_DEBUG environment variable is set.
// It adds a stacktrace to the log message.
func debugF(format string, a ...interface{}) {
	if os.Getenv("DAGGER_DEBUG") != "" {
		format = fmt.Sprintf("DEBUG: %s\n", format)
		log.Printf(format, a...)
	}
}
