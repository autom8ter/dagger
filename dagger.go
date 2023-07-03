package dagger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"math"
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
	var visited = map[string]struct{}{}
	q := NewBlockingQueue[*searchItem[N]](0)
	ctx, cancel := context.WithCancel(ctx)
	mu := &sync.RWMutex{}
	var errChan = make(chan error, 1)
	defer cancel()
	go func() {
		egp, ctx := errgroup.WithContext(ctx)
		if err := g.bfs(ctx, &bfsState[N]{
			visited:      visited,
			egp:          egp,
			mu:           mu,
			q:            q,
			reverse:      reverse,
			root:         start,
			next:         start,
			relationship: "",
		}); err != nil {
			errChan <- err
		}
		if err := egp.Wait(); err != nil {
			errChan <- err
		}
		cancel()
	}()
	q.RangeContext(ctx, func(element *searchItem[N]) bool {
		if element.node.Value.ID() != start.Value.ID() {
			return search(ctx, element.relationship, element.node)
		}
		return true
	})
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// DFS executes a depth first search on the graph starting from the current node.
// The reverse parameter determines whether the search is reversed or not.
// The fn parameter is a function that is called on each node in the graph. If the function returns false, the search is stopped.
func (g *DirectedGraph[N]) DFS(ctx context.Context, reverse bool, start *GraphNode[N], fn GraphSearchFunc[N]) error {
	var visited = map[string]struct{}{}
	stack := NewStack[*searchItem[N]]()
	egp, ctx := errgroup.WithContext(ctx)
	if err := g.dfs(ctx, &dfsState[N]{
		visited:      visited,
		egp:          egp,
		mu:           &sync.RWMutex{},
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
			return fn(ctx, element.relationship, element.node)
		}
		return true
	})
	return nil
}

// searchItem is an item that is used in the search queue/stack in DFS/BFS
type searchItem[N Node] struct {
	node         *GraphNode[N]
	relationship string
}

type dfsState[N Node] struct {
	visited      map[string]struct{}
	egp          *errgroup.Group
	mu           *sync.RWMutex
	stack        *Stack[*searchItem[N]]
	reverse      bool
	root         *GraphNode[N]
	next         *GraphNode[N]
	relationship string
}

func (g *DirectedGraph[N]) dfs(ctx context.Context, state *dfsState[N]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	if _, ok := state.visited[state.next.Value.ID()]; !ok {
		state.visited[state.next.Value.ID()] = struct{}{}
		state.stack.Push(&searchItem[N]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				g.dfs(ctx, &dfsState[N]{
					egp:          state.egp,
					mu:           state.mu,
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.To,
					relationship: edge.Relationship,
					visited:      state.visited,
				})
				state.mu.RLock()
				defer state.mu.RUnlock()
				return len(state.visited) < g.nodes.Len() && ctx.Err() == nil
			})
		} else {
			state.next.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
				g.dfs(ctx, &dfsState[N]{
					egp:          state.egp,
					mu:           state.mu,
					stack:        state.stack,
					reverse:      state.reverse,
					root:         state.root,
					next:         edge.From,
					relationship: edge.Relationship,
					visited:      state.visited,
				})
				state.mu.RLock()
				defer state.mu.RUnlock()
				return len(state.visited) < g.nodes.Len() && ctx.Err() == nil
			})
		}
	}
	return ctx.Err()
}

type bfsState[N Node] struct {
	visited      map[string]struct{}
	egp          *errgroup.Group
	mu           *sync.RWMutex
	q            *BlockingQueue[*searchItem[N]]
	reverse      bool
	root         *GraphNode[N]
	next         *GraphNode[N]
	relationship string
}

func (g *DirectedGraph[N]) bfs(ctx context.Context, state *bfsState[N]) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if state.next == nil {
		state.next = state.root
	}
	state.mu.Lock()
	_, ok := state.visited[state.next.Value.ID()]
	if !ok {
		state.visited[state.next.Value.ID()] = struct{}{}
	}
	state.mu.Unlock()
	if !ok {
		state.q.Push(&searchItem[N]{
			node:         state.next,
			relationship: state.relationship,
		})
		if state.reverse {
			state.egp.Go(func() error {
				state.next.edgesTo.Range(func(key string, edge *GraphEdge[N]) bool {
					g.bfs(ctx, &bfsState[N]{
						visited:      state.visited,
						egp:          state.egp,
						mu:           state.mu,
						q:            state.q,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.From,
						relationship: edge.Relationship,
					})
					state.mu.RLock()
					defer state.mu.RUnlock()
					return len(state.visited) < g.nodes.Len() && ctx.Err() == nil
				})
				return nil
			})
		} else {
			state.egp.Go(func() error {
				state.next.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
					g.bfs(ctx, &bfsState[N]{
						visited:      state.visited,
						egp:          state.egp,
						mu:           state.mu,
						q:            state.q,
						reverse:      state.reverse,
						root:         state.root,
						next:         edge.To,
						relationship: edge.Relationship,
					})
					state.mu.RLock()
					defer state.mu.RUnlock()
					return len(state.visited) < g.nodes.Len() && ctx.Err() == nil
				})
				return nil
			})
		}
	}
	return ctx.Err()
}

// Acyclic returns true if the graph contains no cycles.
func (g *DirectedGraph[N]) Acyclic() bool {
	isAcyclic := false
	g.nodes.Range(func(key string, node *GraphNode[N]) bool {
		if node.edgesFrom.Len() > 0 {
			visited := map[string]struct{}{}
			onStack := map[string]bool{}
			if g.isAcyclic(node, visited, onStack) {
				isAcyclic = true
				return false
			}
		}
		return true
	})
	return isAcyclic
}

// isAcyclic returns true if the graph contains no cycles.
func (g *DirectedGraph[N]) isAcyclic(node *GraphNode[N], visited map[string]struct{}, onStack map[string]bool) bool {
	visited[node.Value.ID()] = struct{}{}
	onStack[node.Value.ID()] = true
	result := false
	node.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
		if _, ok := visited[edge.To.Value.ID()]; !ok {
			if g.isAcyclic(edge.To, visited, onStack) {
				result = true
				return false
			}
		} else if onStack[edge.To.Value.ID()] {
			result = true
			return false
		}
		return true
	})
	return result
}

// StronglyConnected returns the strongly connected components of the graph using Tarjan's algorithm.
func (g *DirectedGraph[N]) StronglyConnected() ([][]*GraphNode[N], error) {
	if !g.Acyclic() {
		return nil, fmt.Errorf("Strongly connected components can only be computed in acyclic graphs")
	}

	state := &sccState[N]{
		components: make([][]*GraphNode[N], 0),
		stack:      NewStack[*GraphNode[N]](),
		onStack:    make(map[string]bool),
		visited:    make(map[string]struct{}),
		lowlink:    make(map[string]int),
		index:      make(map[string]int),
	}
	g.nodes.Range(func(key string, node *GraphNode[N]) bool {
		if _, ok := state.visited[key]; !ok {
			g.stronglyConnected(node, state)
		}
		return true
	})
	return state.components, nil
}

type sccState[N Node] struct {
	components [][]*GraphNode[N]
	stack      *Stack[*GraphNode[N]]
	onStack    map[string]bool
	visited    map[string]struct{}
	lowlink    map[string]int
	index      map[string]int
	time       int
}

func (g *DirectedGraph[N]) stronglyConnected(node *GraphNode[N], state *sccState[N]) {
	nodeID := node.Value.ID()
	state.stack.Push(node)

	state.onStack[nodeID] = true
	state.visited[nodeID] = struct{}{}
	state.index[nodeID] = state.time
	state.lowlink[nodeID] = state.time

	state.time++
	g.nodes.Range(func(key string, adjacency *GraphNode[N]) bool {
		if _, ok := state.visited[adjacency.Value.ID()]; !ok {
			g.stronglyConnected(adjacency, state)

			smallestLowlink := math.Min(
				float64(state.lowlink[nodeID]),
				float64(state.lowlink[adjacency.Value.ID()]),
			)
			state.lowlink[nodeID] = int(smallestLowlink)
		} else {
			// If the adjacent vertex already is on the stack, the edge joining
			// the current and the adjacent vertex is a back ege. Therefore, the
			// lowlink value of the vertex has to be updated to the index of the
			// adjacent vertex if it is smaller than the current lowlink value.
			if state.onStack[adjacency.Value.ID()] {
				smallestLowlink := math.Min(
					float64(state.lowlink[nodeID]),
					float64(state.index[adjacency.Value.ID()]),
				)
				state.lowlink[nodeID] = int(smallestLowlink)
			}
		}
		return true
	})

	// If the lowlink value of the vertex is equal to its DFS value, this is the
	// head vertex of a strongly connected component that's shaped by the vertex
	// and all vertices on the stack.
	if state.lowlink[nodeID] == state.index[nodeID] {
		var id string
		var component []*GraphNode[N]

		for id != nodeID {
			val, _ := state.stack.Pop()
			id = val.Value.ID()
			state.onStack[id] = false

			component = append(component, val)
		}
		state.components = append(state.components, component)
	}
}

// ShortestPath returns the shortest path from source to target using Dijkstra's algorithm.
func (g *DirectedGraph[N]) ShortestPath(source, target *GraphNode[N]) ([]*GraphNode[N], error) {
	weights := make(map[string]float64)
	visited := make(map[string]bool)

	weights[source.Value.ID()] = 0
	visited[target.Value.ID()] = true

	queue := NewPriorityQueue[*GraphNode[N]]()

	g.nodes.Range(func(key string, node *GraphNode[N]) bool {
		if key != source.Value.ID() {
			weights[key] = math.Inf(1)
			visited[key] = false
		}
		queue.Push(node, weights[key])
		return true
	})
	cheapestParent := make(map[*GraphNode[N]]*GraphNode[N])

	for queue.Len() > 0 {
		vertex, _ := queue.Pop()
		hasInfiniteWeight := math.IsInf(weights[vertex.Value.ID()], 1)
		g.nodes.Range(func(key string, adjacency *GraphNode[N]) bool {
			adjacency.edgesFrom.Range(func(key string, edge *GraphEdge[N]) bool {
				edgeWeight := edge.Metadata["weight"]
				if edgeWeight == "" {
					edgeWeight = "1"
				}
				fl, _ := strconv.ParseFloat(edgeWeight, 64)

				weight := weights[vertex.Value.ID()] + fl

				if weight < weights[adjacency.Value.ID()] && !hasInfiniteWeight {
					weights[adjacency.Value.ID()] = weight
					cheapestParent[adjacency] = vertex
					queue.UpdatePriority(adjacency, weight)
				}
				return true
			})
			return true
		})
	}

	path := []*GraphNode[N]{target}
	current := target

	for current != source {
		if _, ok := cheapestParent[current]; !ok {
			return nil, fmt.Errorf("no path from %s to %s", source.Value.ID(), target.Value.ID())
		}
		current = cheapestParent[current]
		path = append([]*GraphNode[N]{current}, path...)
	}

	return path, nil
}

// PredecessorMap returns a map of all predecessors of all nodes in the graph.
func (d *DirectedGraph[N]) PredecessorMap() (map[*GraphNode[N]]map[*GraphNode[N]]*GraphEdge[N], error) {
	m := make(map[*GraphNode[N]]map[*GraphNode[N]]*GraphEdge[N], d.nodes.Len())
	d.nodes.Range(func(id string, node *GraphNode[N]) bool {
		m[node] = make(map[*GraphNode[N]]*GraphEdge[N])
		return true
	})
	d.edges.Range(func(id string, edge *GraphEdge[N]) bool {
		if _, ok := m[edge.To]; !ok {
			m[edge.To] = make(map[*GraphNode[N]]*GraphEdge[N])
		}
		m[edge.To][edge.From] = edge
		return true
	})
	return m, nil
}

func (g *DirectedGraph[N]) TopologicalSort() ([]*GraphNode[N], error) {
	if !g.Acyclic() {
		return nil, fmt.Errorf("topological sort cannot be computed on cyclical graph")
	}

	count := g.nodes.Len()

	predecessorMap, err := g.PredecessorMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get predecessor map: %w", err)
	}

	queue := NewQueue[*GraphNode[N]]()

	for vertex, predecessors := range predecessorMap {
		if len(predecessors) == 0 {
			queue.Push(vertex)
		}
	}

	order := make([]*GraphNode[N], 0, count)
	visited := make(map[string]struct{}, count)
	queue.Range(func(vertex *GraphNode[N]) bool {
		currentVertex, ok := queue.Pop()
		if !ok {
			return false
		}
		if _, ok := visited[currentVertex.Value.ID()]; ok {
			return true
		}

		order = append(order, currentVertex)
		visited[currentVertex.Value.ID()] = struct{}{}

		for vertex, predecessors := range predecessorMap {
			delete(predecessors, currentVertex)

			if len(predecessors) == 0 {
				queue.Push(vertex)
			}
		}
		return true
	})

	if len(order) != count {
		return nil, fmt.Errorf("topological sort cannot be computed on graph with cycles %d != %d", len(order), count)
	}

	return order, nil
}

func (g *DirectedGraph[N]) topology(reverse bool, stack *Stack[string], node *GraphNode[N], permanent, temporary map[string]struct{}) {
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

// BlockingQueue is a basic FIFO BlockingQueue based on a buffered channel
type BlockingQueue[T any] struct {
	closeOnce sync.Once
	ch        chan T
}

// NewBlockingQueue returns a new BlockingQueue with the given max size. When the max size is reached, the queue will block until a value is removed.
// If maxSize is 0, the queue will always block until a value is removed.
func NewBlockingQueue[T any](maxSize int) *BlockingQueue[T] {
	vals := make(chan T, maxSize)
	return &BlockingQueue[T]{ch: vals}
}

// Range executes a provided function once for each BlockingQueue element until it returns false.
func (q *BlockingQueue[T]) Range(fn func(element T) bool) {
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

// RangeContext executes a provided function once for each BlockingQueue element until it returns false or the context is done.
func (q *BlockingQueue[T]) RangeContext(ctx context.Context, fn func(element T) bool) {
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

// Close closes the BlockingQueue channel.
func (q *BlockingQueue[T]) Close() {
	q.closeOnce.Do(func() {
		close(q.ch)
	})
}

// Push adds an element to the end of the BlockingQueue.
func (q *BlockingQueue[T]) Push(val T) {
	q.ch <- val
}

// Pop removes and returns an element from the beginning of the BlockingQueue.
func (q *BlockingQueue[T]) Pop() (T, bool) {
	select {
	case r, ok := <-q.ch:
		return r, ok
	}
}

// Len returns the number of elements in the BlockingQueue.
func (q *BlockingQueue[T]) Len() int {
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

// Range executes a provided function once for each Queue element until it returns false.
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
