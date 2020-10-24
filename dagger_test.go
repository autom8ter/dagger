package dagger_test

import (
	"fmt"
	"github.com/autom8ter/dagger"
	"testing"
	"time"
)

var (
	coleman = dagger.NewNode("user", fmt.Sprintf("cword_%v", time.Now().UnixNano()), map[string]interface{}{
		"name": "coleman",
	})
	tyler = dagger.NewNode("user", fmt.Sprintf("twash_%v", time.Now().UnixNano()), map[string]interface{}{
		"name": "tyler",
	})
	sarah = dagger.NewNode("user", fmt.Sprintf("swash_%v", time.Now().UnixNano()), map[string]interface{}{
		"name": "sarah",
	})
	lacee = dagger.NewNode("user", fmt.Sprintf("ljans_%v", time.Now().UnixNano()), map[string]interface{}{
		"name": "lacee",
	})
	// random id will be generated if one isn't provided
	charlie = dagger.NewNode("dog", "", map[string]interface{}{
		"name":   "charlie",
		"weight": 25,
	})
)

func seedT(t *testing.T) {
	if coleman.GetString("name") != "coleman" {
		exit("expected name to be coleman")
	}
	if err := coleman.Connect(tyler, "friend", true); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(lacee, "friend", true); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(lacee, "fiance", true); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(sarah, "wife", true); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(charlie, "pet", false); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(charlie, "pet", false); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(lacee, "owner", false); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(coleman, "owner", false); err != nil {
		t.Fatal(err)
	}
	charlie.Patch(map[string]interface{}{
		"weight": 19,
	})
	if charlie.GetInt("weight") != 19 {
		t.Fatal("expected charlie's weight to be 19!")
	}
	// check to make sure edge is patched
	coleman.EdgesFrom(func(e *dagger.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			if e.To().GetInt("weight") != 19 {
				t.Fatal("failed to patch charlie's weight")
			}
		}
		return true
	})
	// remove from graph
	charlie.Remove()
	// no longer in graph
	if dagger.HasNode(charlie) {
		t.Fatal("failed to delete node - (charlie)")
	}
	// check to make sure edge no longer exists(cascade)
	coleman.EdgesFrom(func(e *dagger.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			t.Fatal("failed to delete node - (charlie)")
		}
		return true
	})
	// check to make sure edge no longer exists(cascade)
	lacee.EdgesFrom(func(e *dagger.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			t.Fatal("failed to delete node - (charlie)")
		}
		return true
	})
}

func seedB(t *testing.B) {
	if err := coleman.Connect(tyler, "friend", true); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(lacee, "friend", true); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(lacee, "fiance", true); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(sarah, "wife", true); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(charlie, "pet", false); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(charlie, "pet", false); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(lacee, "owner", false); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(coleman, "owner", false); err != nil {
		t.Fatal(err)
	}
	charlie.Patch(map[string]interface{}{
		"weight": 19,
	})
	coleman.EdgesFrom(func(e *dagger.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			if e.To().GetInt("weight") != 19 {
				t.Fatal("failed to patch charlie's weight")
			}
		}
		return true
	})
}

func Test(t *testing.T) {
	seedT(t)
	t.Logf("registered node types = %v\n", dagger.NodeTypes())
	t.Logf("registered edge types = %v\n", dagger.EdgeTypes())
	dagger.RangeNodes(func(n *dagger.Node) bool {
		bits, err := n.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bits))
		n.EdgesFrom(func(e *dagger.Edge) bool {
			bits, err := e.JSON()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(bits))
			return true
		})
		n.EdgesTo(func(e *dagger.Edge) bool {
			bits, err := e.JSON()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(bits))
			return true
		})
		return true
	})
}

func Benchmark(t *testing.B) {
	t.ReportAllocs()
	nodes := 0
	edgesFrom := 0
	edgesTo := 0
	//seedB(t)
	t.ResetTimer()
	for n := 0; n < t.N; n++ {
		seedB(t)
		dagger.RangeNodes(func(n *dagger.Node) bool {
			nodes++
			t.Logf("nodes(%v)", nodes)
			n.EdgesFrom(func(e *dagger.Edge) bool {
				edgesFrom++
				t.Logf("edgesFrom(%v)", edgesFrom)
				return true
			})
			n.EdgesTo(func(e *dagger.Edge) bool {
				edgesTo++
				t.Logf("edgesTo(%v)", edgesTo)
				return true
			})
			return true
		})

	}
	t.Logf("visited: %v nodes", nodes)
	t.Logf("visited: %v edgesFrom", edgesFrom)
	t.Logf("visited: %v edgesTo", edgesTo)
}
