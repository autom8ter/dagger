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
	tyler.SetID("twash")
	tyler.SetType("user")
	tyler.SetAll(map[string]interface{}{
		"name": "tyler",
	})
	if err := tyler.Validate(); err != nil {
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
	lacee.SetID("ljanss")
	lacee.SetType("user")
	lacee.SetAll(map[string]interface{}{
		"name": "lacee",
	})
	if err := lacee.Validate(); err != nil {
		t.Fatal(err)
	}
	bits, err := coleman.JSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bits))
}
