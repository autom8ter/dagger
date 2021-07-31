package dagger_test

import (
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/util"
	"testing"
)

func Test(t *testing.T) {
	var g = dagger.NewGraph()
	vm1 := g.SetNode(dagger.Path{
		XID:   "virtual_machine_1",
		XType: "infra",
	}, map[string]interface{}{})
	vm2 := g.SetNode(dagger.Path{
		XID:   "virtual_machine_2",
		XType: "infra",
	}, map[string]interface{}{})
	k8s := g.SetNode(dagger.Path{
		XID:   "kubernetes",
		XType: "infra",
	}, map[string]interface{}{})
	redis := g.SetNode(dagger.Path{
		XID:   "redis",
		XType: "infra",
	}, map[string]interface{}{
		"port": "6379",
	})
	mongo := g.SetNode(dagger.Path{
		XID:   "mongo",
		XType: "infra",
	}, map[string]interface{}{
		"port": "5568",
	})
	httpServer := g.SetNode(dagger.Path{
		XID:   "http",
		XType: "infra",
	}, map[string]interface{}{
		"port": "8080",
	})
	g.RangeNodes("*", func(n dagger.Node) bool {
		t.Logf("found node in graph: %s.%s\n", n.XType, n.XID)
		return true
	})
	_, err := g.SetEdge(k8s.Path, vm1.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(k8s.Path, vm2.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(mongo.Path, redis.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(httpServer.Path, k8s.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(redis.Path, k8s.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(mongo.Path, k8s.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(httpServer.Path, redis.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(httpServer.Path, mongo.Path, dagger.Node{
		Path: dagger.Path{
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(g.String())
	//g.EdgesFrom("*", httpServer.Path, func(e dagger.Edge) bool {
	//	fmt.Println(util.JSONString(e))
	//	return true
	//})
	i := 0
	g.RangeEdgesFrom("*", httpServer.Path, func(e dagger.Edge) bool {
		i++
		return true
	})
	if i != 3 {
		t.Fatal("expected 3", i)
	}
	i = 0
	g.RangeEdgesFrom("*", redis.Path, func(e dagger.Edge) bool {
		i++
		return true
	})
	if i != 1 {
		t.Fatal("expected 1", i)
	}
	i = 0
	g.RangeEdgesFrom("*", k8s.Path, func(e dagger.Edge) bool {
		i++
		return true
	})
	if i != 2 {
		t.Fatal("expected 2", i)
	}
	i = 0
	g.RangeEdgesTo("*", k8s.Path, func(e dagger.Edge) bool {
		i++
		return true
	})
	if i != 3 {
		t.Fatal("expected 3", i)
	}
	i = 0
	g.ReverseDFS("depends_on", httpServer.Path, func(node dagger.Node) bool {
		i++
		t.Logf("(%v) Reverse DFS: %s", i, util.JSONString(node))
		return true
	})
	if i > 0 {
		t.Fatal("expected 0", i)
	}
	i = 0
	g.ReverseBFS("depends_on", httpServer.Path, func(node dagger.Node) bool {
		i++
		t.Logf("(%v) Reverse BFS: %s", i, util.JSONString(node))
		return true
	})
	if i > 0 {
		t.Fatal("expected 0", i)
	}

	i = 0
	g.TopologicalSort("infra", "depends_on", func(node dagger.Node) bool {
		i++
		t.Logf("(%v) TopologicalSort: %s", i, util.JSONString(node))
		return true
	})
	if i != 6 {
		t.Fatal("expected 6", i)
	}

	i = 0
	g.ReverseTopologicalSort("infra", "depends_on", func(node dagger.Node) bool {
		i++
		t.Logf("(%v) ReverseTopologicalSort: %s", i, util.JSONString(node))
		return true
	})
	if i != 6 {
		t.Fatal("expected 6", i)
	}

	//g.BFS("*", mongo.Path, func(node dagger.Node) bool {
	//	bits, _ := json.MarshalIndent(&node, "", "    ")
	//	t.Log("BFS: ", string(bits))
	//	return true
	//})
}
