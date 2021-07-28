package dagger_test

import (
	"fmt"
	"github.com/autom8ter/dagger"
)

func ExampleGraph_SetNode() {
	var g = dagger.NewGraph()
	g.SetNode(dagger.Path{
		XID:   "virtual_machine_1",
		XType: "infra",
	}, map[string]interface{}{})
	g.SetNode(dagger.Path{
		XID:   "virtual_machine_2",
		XType: "infra",
	}, map[string]interface{}{})
	g.SetNode(dagger.Path{
		XID:   "kubernetes",
		XType: "infra",
	}, map[string]interface{}{})
	g.SetNode(dagger.Path{
		XID:   "redis",
		XType: "infra",
	}, map[string]interface{}{
		"port": "6379",
	})
	g.SetNode(dagger.Path{
		XID:   "mongo",
		XType: "infra",
	}, map[string]interface{}{
		"port": "5568",
	})
	g.SetNode(dagger.Path{
		XID:   "http",
		XType: "infra",
	}, map[string]interface{}{
		"port": "8080",
	})
	g.RangeNodes("*", func(n dagger.Node) bool {
		fmt.Printf("found node in graph: %s.%s\n", n.XType, n.XID)
		return true
	})
}
