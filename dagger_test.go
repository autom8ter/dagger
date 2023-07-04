package dagger_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/autom8ter/dagger/v3"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("DAGGER_DEBUG", "true")
}

func TestGraph(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.Run("set node", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		assert.NotNil(t, node)
	})
	t.Run("set edge", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
	})
	t.Run("set edge with node", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
	})
	t.Run("set edge with node then get", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		edge, ok := graph.GetEdge(edge.ID)
		assert.True(t, ok)
		assert.NotNil(t, edge)
	})
	t.Run("check edges from/to", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		node1.EdgesFrom("", func(e *dagger.GraphEdge[string]) bool {
			assert.Equal(t, e.ID, edge.ID)
			return true
		})
		node1.EdgesTo("", func(e *dagger.GraphEdge[string]) bool {
			assert.NotEqual(t, e.ID, edge.ID)
			return true
		})
		node2.EdgesTo("", func(e *dagger.GraphEdge[string]) bool {
			assert.Equal(t, e.ID, edge.ID)
			return true
		})
		node2.EdgesFrom("", func(e *dagger.GraphEdge[string]) bool {
			assert.NotEqual(t, e.ID, edge.ID)
			return true
		})
	})
	t.Run("remove node", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		assert.NoError(t, node1.Remove())
		_, ok := graph.GetNode(node1.ID())
		assert.False(t, ok)
		_, ok = graph.GetEdge(edge.ID)
		assert.False(t, ok)
	})
	t.Run("remove edge", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		node1.RemoveEdge(edge.ID)
		_, ok := graph.GetEdge(edge.ID)
		assert.False(t, ok)

	})
	t.Run("graphviz", func(t *testing.T) {
		graph, err := dagger.NewDAG[string](dagger.WithVizualization())
		assert.NoError(t, err)
		node1 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		node2 := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		edge, err := node1.SetEdge(node2, "connected", map[string]string{})
		assert.Nil(t, err)
		assert.NotNil(t, edge)
		img, err := graph.GraphViz()
		assert.NoError(t, err)
		assert.NotNil(t, img)
	})
	t.Run("breadthFirstSearch", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		for i := 0; i < 100; i++ {
			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.ID())
		nodes := make([]*dagger.GraphNode[string], 0)
		assert.NoError(t, graph.BFS(ctx, false, lastNode, func(ctx context.Context, relationship string, node *dagger.GraphNode[string]) bool {
			nodes = append(nodes, node)
			return true
		}))

		assert.Equal(t, 100, len(nodes))
	})
	t.Run("depthFirstSearch", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		for i := 0; i < 100; i++ {
			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.ID())
		nodes := make([]*dagger.GraphNode[string], 0)

		assert.NoError(t, graph.DFS(ctx, false, lastNode, func(ctx context.Context, relationship string, node *dagger.GraphNode[string]) bool {
			nodes = append(nodes, node)
			return true
		}))
		assert.Equal(t, 100, len(nodes))
	})
	t.Run("acyclic", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		for i := 0; i < 100; i++ {
			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
		}
		nc, ec := graph.Size()
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.ID())
		assert.True(t, graph.Acyclic())
	})
	t.Run("topological reverse sort", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		for i := 0; i < 100; i++ {
			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
			lastNode = node
		}
		nc, ec := graph.Size()
		assert.True(t, graph.Acyclic())
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.ID())
		nodes, err := graph.TopologicalSort(true)
		assert.NoError(t, err)
		var n *dagger.GraphNode[string]
		for _, node := range nodes {
			if n == nil {
				n = node
				continue
			}
			split := strings.Split(n.ID(), "-")
			lastNodeID, _ := strconv.Atoi(split[len(split)-1])
			split = strings.Split(node.ID(), "-")
			nodeID, _ := strconv.Atoi(split[len(split)-1])
			assert.LessOrEqual(t, lastNodeID, nodeID)
			n = node
		}
		assert.Equal(t, 101, len(nodes))
	})
	t.Run("topological sort", func(t *testing.T) {
		graph, err := dagger.NewDAG[string]()
		assert.NoError(t, err)
		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
		for i := 0; i < 100; i++ {
			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
				"weight": fmt.Sprintf("%d", i),
			})
			assert.Nil(t, err)
			assert.NotNil(t, edge)
			lastNode = node
		}
		nc, ec := graph.Size()
		assert.True(t, graph.Acyclic())
		t.Logf("nodes=%v edges=%v last node: %s", nc, ec, lastNode.ID())
		nodes, err := graph.TopologicalSort(false)
		assert.NoError(t, err)
		var n *dagger.GraphNode[string]
		for _, node := range nodes {
			if n == nil {
				n = node
				continue
			}
			split := strings.Split(n.ID(), "-")
			lastNodeID, _ := strconv.Atoi(split[len(split)-1])
			split = strings.Split(node.ID(), "-")
			nodeID, _ := strconv.Atoi(split[len(split)-1])
			assert.GreaterOrEqual(t, lastNodeID, nodeID)
			n = node
		}
		assert.Equal(t, 101, len(nodes))
	})
	//t.Run("strongly connected", func(t *testing.T) {
	//	graph, err := dagger.NewDAG[string]()
	//	{
	//		lastNode := graph.SetNode(dagger.UniqueID("node"), "", map[string]string{})
	//		for i := 0; i < 50; i++ {
	//			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
	//			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
	//				"weight": fmt.Sprintf("%d", i),
	//			})
	//			assert.Nil(t, err)
	//			assert.NotNil(t, edge)
	//			lastNode = node
	//		}
	//	}
	//	{
	//		lastNode := graph.SetNode(dagger.UniqueID("node1"))
	//		for i := 51; i < 100; i++ {
	//			node := graph.SetNode(fmt.Sprintf("node-%d", i), "", map[string]string{})
	//			edge, err := lastNode.SetEdge(node, "connected", map[string]string{
	//				"weight": fmt.Sprintf("%d", i),
	//			})
	//			assert.Nil(t, err)
	//			assert.NotNil(t, edge)
	//			lastNode = node
	//		}
	//	}
	//	assert.True(t, graph.Acyclic())
	//	components, err := graph.StronglyConnected()
	//	assert.NoError(t, err)
	//	assert.Equal(t, 2, len(components))
	//	bits, _ := json.MarshalIndent(components, "", "  ")
	//	t.Logf("strongly connected components: %v", string(bits))
	//})
}

func TestBoundedQueue(t *testing.T) {
	q := dagger.NewBoundedQueue[string](100)
	for i := 0; i < 100; i++ {
		q.Push(string(fmt.Sprintf("node-%d", i)))
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
	s := dagger.NewStack[string]()
	for i := 0; i < 100; i++ {
		s.Push(string(fmt.Sprintf("node-%d", i)))
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
	s := dagger.NewSet[string]()
	for i := 0; i < 100; i++ {
		s.Add(string(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, s.Len(), 100)
	for i := 0; i < 100; i++ {
		ok := s.Contains(string(fmt.Sprintf("node-%d", i)))
		assert.True(t, ok)
	}
	assert.Equal(t, s.Len(), 100)
	for i := 0; i < 100; i++ {
		s.Remove(string(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, s.Len(), 0)
}

func TestNewPriorityQueue(t *testing.T) {
	q := dagger.NewPriorityQueue[string]()
	for i := 0; i < 100; i++ {
		q.Push(string(fmt.Sprintf("node-%d", i)), float64(i))
	}
	assert.Equal(t, q.Len(), 100)
	first, _ := q.Peek()
	assert.EqualValues(t, first, string("node-0"))
	for i := 0; i < 100; i++ {
		v, ok := q.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, q.Len(), 0)
}

func TestQueue(t *testing.T) {
	q := dagger.NewQueue[string]()
	for i := 0; i < 100; i++ {
		q.Push(string(fmt.Sprintf("node-%d", i)))
	}
	assert.Equal(t, q.Len(), 100)
	for i := 0; i < 100; i++ {
		v, ok := q.Pop()
		assert.NotNil(t, v)
		assert.True(t, ok)
	}
	assert.Equal(t, q.Len(), 0)
}

func TestNewChannelGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	g := dagger.NewChannelGroup[string](ctx)
	wg := sync.WaitGroup{}
	count := int64(0)
	for i := 0; i < 100; i++ {
		ch := g.Channel(ctx)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for value := range ch {
				assert.NotNil(t, value)
				t.Logf("(%v) value: %v", i, value)
				atomic.AddInt64(&count, 1)
			}
			t.Logf("(%v) channel closed", i)
		}(i)
	}
	assert.Equal(t, g.Len(), 100)
	for i := 0; i < 100; i++ {
		g.Send(ctx, fmt.Sprintf("node-%d", i))
	}
	time.Sleep(1 * time.Second)
	t.Logf("sent values")
	g.Close()
	t.Logf("closed group")
	wg.Wait()
	assert.Equal(t, g.Len(), 0)
	assert.Equal(t, count, int64(10000))
}
