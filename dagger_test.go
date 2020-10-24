package dagger_test

import (
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/primitive"
	"testing"
	"time"
)

func seedT(t *testing.T) {
	var (
		coleman = dagger.NewNode("user", fmt.Sprintf("cword_%v", time.Now().UnixNano()), map[string]interface{}{
			"name": "coleman",
		})
		tyler   = dagger.NewNode("user", fmt.Sprintf("twash_%v", time.Now().UnixNano()), map[string]interface{}{
			"name": "tyler",
		})
		sarah   = dagger.NewNode("user", fmt.Sprintf("swash_%v", time.Now().UnixNano()), map[string]interface{}{
			"name": "sarah",
		})
		lacee   = dagger.NewNode("user", fmt.Sprintf("ljans_%v", time.Now().UnixNano()), map[string]interface{}{
			"name": "lacee",
		})
		charlie   = dagger.NewNode("dog", "", map[string]interface{}{
			"name": "charlie",
			"weight": 25,
		})
	)
	if err := coleman.Connect(tyler, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(coleman, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(lacee, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(sarah, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(lacee, "fiance"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(coleman, "fiance"); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(lacee, "wife"); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(tyler, "wife"); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(charlie, "pet"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(charlie, "pet"); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(lacee, "owner"); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(coleman, "owner"); err != nil {
		t.Fatal(err)
	}
}

func seedB(t *testing.B) {
	var (
		coleman = dagger.NewNode("user", "cword", map[string]interface{}{
			"name": "coleman",
		})
		tyler   = dagger.NewNode("user", "twash", map[string]interface{}{
			"name": "tyler",
		})
		sarah   = dagger.NewNode("user", "swash", map[string]interface{}{
			"name": "sarah",
		})
		lacee   = dagger.NewNode("user", "ljans", map[string]interface{}{
			"name": "lacee",
		})
		charlie   = dagger.NewNode("dog", "", map[string]interface{}{
			"name": "charlie",
			"weight": 25,
		})
	)
	if err := coleman.Connect(tyler, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(coleman, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(lacee, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(sarah, "friend"); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(lacee, "fiance"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(coleman, "fiance"); err != nil {
		t.Fatal(err)
	}
	if err := tyler.Connect(lacee, "wife"); err != nil {
		t.Fatal(err)
	}
	if err := sarah.Connect(tyler, "wife"); err != nil {
		t.Fatal(err)
	}
	if err := coleman.Connect(charlie, "pet"); err != nil {
		t.Fatal(err)
	}
	if err := lacee.Connect(charlie, "pet"); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(lacee, "owner"); err != nil {
		t.Fatal(err)
	}
	if err := charlie.Connect(coleman, "owner"); err != nil {
		t.Fatal(err)
	}
}

func Test(t *testing.T) {
	seedT(t)
	t.Log(dagger.NodeTypes())
	t.Log(dagger.EdgeTypes())
	dagger.RangeNodes(func(n *dagger.Node) bool {
		bits, err := n.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bits) )
		n.EdgesFrom(func(e *primitive.Edge) bool {
			bits, err := e.JSON()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(bits) )
			return true
		})
		n.EdgesTo(func(e *primitive.Edge) bool {
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
			n.EdgesFrom(func(e *primitive.Edge) bool {
				edgesFrom++
				t.Logf("here")
				return true
			})
			n.EdgesTo(func(e *primitive.Edge) bool {
				edgesTo++
				return true
			})
			return true
		})
	}
	t.Logf("visited: %v nodes", nodes)
	t.Logf("visited: %v edgesFrom", edgesFrom)
	t.Logf("visited: %v edgesTo", edgesTo)
}

