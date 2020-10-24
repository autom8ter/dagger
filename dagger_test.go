package dagger_test

import (
	"github.com/autom8ter/dagger"
	"testing"
)

func Test(t *testing.T) {
	var (
		coleman = dagger.Node{}
		tyler   = dagger.Node{}
		sarah   = dagger.Node{}
		lacee   = dagger.Node{}
	)
	coleman.SetID("cword")
	coleman.SetType("user")
	coleman.SetAll(map[string]interface{}{
		"name": "coleman",
	})
	if err := coleman.Validate(); err != nil {
		t.Fatal(err)
	}
	dagger.AddNode(coleman)
	tyler.SetID("twash")
	tyler.SetType("user")
	tyler.SetAll(map[string]interface{}{
		"name": "tyler",
	})
	if err := tyler.Validate(); err != nil {
		t.Fatal(err)
	}
	dagger.AddNode(tyler)
	if err := dagger.AddEdge(&dagger.Edge{
		Node: map[string]interface{}{
			"type":   "friend",
			"source": "school",
		},
		From: coleman,
		To:   tyler,
	}); err != nil {
		t.Fatal(err)
	}
	sarah.SetID("swash")
	sarah.SetType("user")
	sarah.SetAll(map[string]interface{}{
		"name": "sarah",
	})
	if err := sarah.Validate(); err != nil {
		t.Fatal(err)
	}
	dagger.AddNode(sarah)
	if err := dagger.AddEdge(&dagger.Edge{
		Node: map[string]interface{}{
			"type":   "wife",
			"source": "college",
		},
		From: tyler,
		To:   sarah,
	}); err != nil {
		t.Fatal(err)
	}
	lacee.SetID("ljanss")
	lacee.SetType("user")
	lacee.SetAll(map[string]interface{}{
		"name": "lacee",
	})
	if err := lacee.Validate(); err != nil {
		t.Fatal(err)
	}
	dagger.AddNode(lacee)
	if err := dagger.AddEdge(&dagger.Edge{
		Node: map[string]interface{}{
			"type":   "fiance",
			"source": "college",
		},
		From: coleman,
		To:   lacee,
	}); err != nil {
		t.Fatal(err)
	}
	bits, err := coleman.JSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bits))
	dagger.EdgesFrom(coleman, func(e *dagger.Edge) bool {
		bits, err := e.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bits))
		return true
	})
}
