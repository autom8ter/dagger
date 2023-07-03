package dagger_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/autom8ter/dagger"
	"github.com/stretchr/testify/assert"
)

func init() {
	// os.Setenv("DAGGER_DEBUG", "true")
}

func TestGraph(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.Run("set node", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node := graph.SetNode(dagger.UniqueID("node"))
		assert.NotNil(t, node)
	})
	t.Run("set edge", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
	})
	t.Run("set edge with node", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
	})
	t.Run("set edge with node then get", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		edge, ok := graph.GetEdge(edge.ID)
		assert.True(t, ok)
		assert.NotNil(t, edge)
	})
	t.Run("check edges from/to", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		node1.EdgesFrom(func(e *dagger.GraphEdge[dagger.String]) bool {
			assert.Equal(t, e.ID, edge.ID)
			return true
		})
		node1.EdgesTo(func(e *dagger.GraphEdge[dagger.String]) bool {
			assert.NotEqual(t, e.ID, edge.ID)
			return true
		})
		node2.EdgesTo(func(e *dagger.GraphEdge[dagger.String]) bool {
			assert.Equal(t, e.ID, edge.ID)
			return true
		})
		node2.EdgesFrom(func(e *dagger.GraphEdge[dagger.String]) bool {
			assert.NotEqual(t, e.ID, edge.ID)
			return true
		})
	})
	t.Run("remove node", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		assert.NoError(t, node1.Remove())
		_, ok := graph.GetNode(node1.Value.ID())
		assert.False(t, ok)
		_, ok = graph.GetEdge(edge.ID)
		assert.False(t, ok)
	})
	t.Run("remove edge", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		node1.RemoveEdge(edge.ID)
		_, ok := graph.GetEdge(edge.ID)
		assert.False(t, ok)

	})
	t.Run("graphviz", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		node1 := graph.SetNode(dagger.UniqueID("node"))
		node2 := graph.SetNode(dagger.UniqueID("node"))
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		img, err := graph.GraphViz()
		assert.NoError(t, err)
		assert.NotNil(t, img)
	})
	t.Run("bfs", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		lastNode := graph.SetNode(dagger.UniqueID("node"))
		for i := 0; i < 100; i++ {
			node := graph.SetNode(dagger.String(fmt.Sprintf("node-%d", i)))
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.Value.ID())
		nodes := make([]*dagger.GraphNode[dagger.String], 0)
		assert.NoError(t, graph.BFS(ctx, false, lastNode, func(ctx context.Context, relationship string, node *dagger.GraphNode[dagger.String]) bool {
			nodes = append(nodes, node)
			return true
		}))

		assert.Equal(t, 100, len(nodes))
	})
	t.Run("dfs", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		lastNode := graph.SetNode(dagger.UniqueID("node"))
		for i := 0; i < 100; i++ {
			node := graph.SetNode(dagger.String(fmt.Sprintf("node-%d", i)))
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.Value.ID())
		nodes := make([]*dagger.GraphNode[dagger.String], 0)

		assert.NoError(t, graph.DFS(ctx, false, lastNode, func(ctx context.Context, relationship string, node *dagger.GraphNode[dagger.String]) bool {
			nodes = append(nodes, node)
			return true
		}))
		assert.Equal(t, 100, len(nodes))
	})
	t.Run("topo", func(t *testing.T) {
		graph := dagger.NewGraph[dagger.String]()
		lastNode := graph.SetNode(dagger.UniqueID("node"))
		for i := 0; i < 100; i++ {
			node := graph.SetNode(dagger.String(fmt.Sprintf("node-%d", i)))
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.Value.ID())
		nodes := make([]*dagger.GraphNode[dagger.String], 0)
		nodes, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.Equal(t, 100, len(nodes))
	})
}

func TestBlockingQueue(t *testing.T) {
	q := dagger.NewBlockingQueue[dagger.String](100)
	for i := 0; i < 100; i++ {
		q.Push(dagger.String(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, q.Len(), 100)
	for i := 0; i < 100; i++ {
		v, ok := q.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, q.Len(), 0)
}

func TestStack(t *testing.T) {
	s := dagger.NewStack[dagger.String]()
	for i := 0; i < 100; i++ {
		s.Push(dagger.String(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, s.Len(), 100)
	for i := 0; i < 100; i++ {
		v, ok := s.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, s.Len(), 0)
}

func TestSet(t *testing.T) {
	s := dagger.NewSet[dagger.String]()
	for i := 0; i < 100; i++ {
		s.Add(dagger.String(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, s.Len(), 100)
	for i := 0; i < 100; i++ {
		ok := s.Contains(dagger.String(fmt.Sprintf("node-%d", i)))
		assert.True(t, ok)
	}
	assert.Equal(t, s.Len(), 100)
	for i := 0; i < 100; i++ {
		s.Remove(dagger.String(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, s.Len(), 0)
}

func TestNewPriorityQueue(t *testing.T) {
	q := dagger.NewPriorityQueue[dagger.String]()
	for i := 0; i < 100; i++ {
		q.Push(dagger.String(fmt.Sprintf("node-%d", i)), float64(i))
	}
	assert.Equal(t, q.Len(), 100)
	first, _ := q.Peek()
	assert.EqualValues(t, first.ID(), dagger.String("node-0"))
	for i := 0; i < 100; i++ {
		v, ok := q.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, q.Len(), 0)
}

func TestQueue(t *testing.T) {
	q := dagger.NewQueue[dagger.String]()
	for i := 0; i < 100; i++ {
		q.Push(dagger.String(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, q.Len(), 100)
	for i := 0; i < 100; i++ {
		v, ok := q.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, q.Len(), 0)
}
