package dagger_test

import (
	"encoding/json"
	"github.com/autom8ter/dagger"
	"testing"
)

func Test(t *testing.T) {
	var g = dagger.NewGraph()

	redis := g.SetNode(dagger.ForeignKey{
		XID:   "redis",
		XType: "infra",
	}, map[string]interface{}{
		"port": "6379",
	})
	mongo := g.SetNode(dagger.ForeignKey{
		XID:   "mongo",
		XType: "infra",
	}, map[string]interface{}{
		"port": "5568",
	})
	httpServer := g.SetNode(dagger.ForeignKey{
		XID:   "http",
		XType: "infra",
	}, map[string]interface{}{
		"port": "8080",
	})
	_, err := g.SetEdge(httpServer, redis, dagger.Node{
		ForeignKey: dagger.ForeignKey{
			XID:   redis.ID(),
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.SetEdge(httpServer, mongo, dagger.Node{
		ForeignKey: dagger.ForeignKey{
			XID:   mongo.ID(),
			XType: "depends_on",
		},
		Attributes: map[string]interface{}{},
	})
	if err != nil {
		t.Fatal(err)
	}

	g.DFS(dagger.AnyType, httpServer, func(node dagger.Node) bool {
		bits, _ := json.MarshalIndent(&node, "", "    ")
		t.Log(string(bits))
		return true
	})
}
