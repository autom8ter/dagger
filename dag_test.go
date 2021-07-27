package dagger_test

import (
	"encoding/json"
	"github.com/autom8ter/dagger"
	"testing"
)

func Test(t *testing.T) {
	var g = dagger.NewGraph()

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
	_, err := g.SetEdge(httpServer.Path, redis.Path, dagger.Node{
		Path: dagger.Path{
			XID:   redis.ID(),
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(httpServer.Path, mongo.Path, dagger.Node{
		Path: dagger.Path{
			XID:   mongo.ID(),
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}

	g.DFS("*", httpServer.Path, func(node dagger.Node) bool {
		bits, _ := json.MarshalIndent(&node, "", "    ")
		t.Log(string(bits))
		return true
	})
}
