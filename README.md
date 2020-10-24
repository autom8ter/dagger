# dagger [![GoDoc](https://godoc.org/github.com/autom8ter/dagger?status.svg)](https://godoc.org/github.com/autom8ter/dagger)

dagger is a blazing fast, concurrency safe, mutable, in-memory directed graph implementation with zero dependencies
    
    import "github.com/autom8ter/dagger"

Design:
- flexibility
- global state
    - see dag package if you want to manage individual graph state manually
- thread safe
- high performance
- simple api

## Example

```go
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

	if err := coleman.Connect(tyler, "friend", true); err != nil {
		exitErr(err)
	}
	if err := sarah.Connect(lacee, "friend", true); err != nil {
		exitErr(err)
	}
	if err := coleman.Connect(lacee, "fiance", true); err != nil {
		exitErr(err)
	}
	if err := tyler.Connect(sarah, "wife", true); err != nil {
		exitErr(err)
	}
	if err := coleman.Connect(charlie, "pet", false); err != nil {
		exitErr(err)
	}
	if err := lacee.Connect(charlie, "pet", false); err != nil {
		exitErr(err)
	}
	if err := charlie.Connect(lacee, "owner", false); err != nil {
		exitErr(err)
	}
	if err := charlie.Connect(coleman, "owner", false); err != nil {
		exitErr(err)
	}
	charlie.Patch(map[string]interface{}{
		"weight": 19,
	})
	if charlie.GetInt("weight") != 19 {
		exit("expected charlie's weight to be 19!")
	}
	// check to make sure edge is patched
	coleman.EdgesFrom(func(e *primitive.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			if e.To.GetInt("weight") != 19 {
				exit("failed to patch charlie's weight")
			}
		}
		return true
	})
	if coleman.GetString("name") != "coleman" {
		exit("expected name to be coleman")
	}
	// remove from graph
	charlie.Remove()
	// no longer in graph
	if dagger.HasNode(charlie) {
		exit("failed to delete node - (charlie)")
	}
	// check to make sure edge no longer exists(cascade)
	coleman.EdgesFrom(func(e *primitive.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			exit("failed to delete node - (charlie)")
		}
		return true
	})
	// check to make sure edge no longer exists(cascade)
	lacee.EdgesFrom(func(e *primitive.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			exit("failed to delete node - (charlie)")
		}
		return true
	})
	fmt.Printf("registered node types = %v", dagger.NodeTypes())
	fmt.Printf("registered edge types = %v", dagger.EdgeTypes())
	dagger.RangeNodes(func(n *dagger.Node) bool {
		bits, err := n.JSON()
		if err != nil {
			exitErr(err)
		}
		fmt.Printf("\nfound node = %v\n", string(bits))
		n.EdgesFrom(func(e *primitive.Edge) bool {
			bits, err := e.JSON()
			if err != nil {
				exitErr(err)
			}
			fmt.Println(string(bits))
			return true
		})
		n.EdgesTo(func(e *primitive.Edge) bool {
			bits, err := e.JSON()
			if err != nil {
				exitErr(err)
			}
			fmt.Println(string(bits))
			return true
		})
		return true
	})

```

Output:
```
    dagger_test.go:136: registered node types = [dog user]
    dagger_test.go:137: registered edge types = [fiance friend owner pet wife]
    dagger_test.go:143: {"_id":"cword_1603572045463636000","_type":"user","name":"coleman"}
    dagger_test.go:149: {"node":{"_id":"9c17980c-e9cc-6e0b-c4a4-81400e2f02e9","_type":"friend"},"from":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"}}
    dagger_test.go:149: {"node":{"_id":"9e39c05a-3113-930a-4ff0-d841e382fbe4","_type":"fiance"},"from":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"}}
    dagger_test.go:149: {"node":{"_id":"fc83d716-c3f8-0548-e3c2-1b4537b3912f","_type":"pet"},"from":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"},"to":{"_id":"8b951d6f-cd73-25f2-fb0a-cd5cb4d93937","_type":"dog","name":"charlie","weight":19}}
    dagger_test.go:157: {"node":{"_id":"9606bd5b-1516-e075-7f3e-7ed2efeee1d2","_type":"fiance"},"from":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"}}
    dagger_test.go:157: {"node":{"_id":"b126a281-78a7-7603-4600-83497b10fec8","_type":"friend"},"from":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"}}
    dagger_test.go:143: {"_id":"twash_1603572045463674000","_type":"user","name":"tyler"}
    dagger_test.go:149: {"node":{"_id":"b126a281-78a7-7603-4600-83497b10fec8","_type":"friend"},"from":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"}}
    dagger_test.go:149: {"node":{"_id":"d9d8df76-e420-a434-5150-862bfd79f075","_type":"wife"},"from":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"}}
    dagger_test.go:157: {"node":{"_id":"9c17980c-e9cc-6e0b-c4a4-81400e2f02e9","_type":"friend"},"from":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"}}
    dagger_test.go:157: {"node":{"_id":"500950c3-1c5a-e1cc-50d8-f307b5614381","_type":"wife"},"from":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"}}
    dagger_test.go:143: {"_id":"swash_1603572045463678000","_type":"user","name":"sarah"}
    dagger_test.go:149: {"node":{"_id":"5150f06e-b979-638e-5dd5-17a480001aee","_type":"friend"},"from":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"}}
    dagger_test.go:149: {"node":{"_id":"500950c3-1c5a-e1cc-50d8-f307b5614381","_type":"wife"},"from":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"}}
    dagger_test.go:157: {"node":{"_id":"bd2166f3-db86-8367-70f9-6053e0ce6b9d","_type":"friend"},"from":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"}}
    dagger_test.go:157: {"node":{"_id":"d9d8df76-e420-a434-5150-862bfd79f075","_type":"wife"},"from":{"_id":"twash_1603572045463674000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"}}
    dagger_test.go:143: {"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"}
    dagger_test.go:149: {"node":{"_id":"bd2166f3-db86-8367-70f9-6053e0ce6b9d","_type":"friend"},"from":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"}}
    dagger_test.go:149: {"node":{"_id":"9606bd5b-1516-e075-7f3e-7ed2efeee1d2","_type":"fiance"},"from":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"}}
    dagger_test.go:149: {"node":{"_id":"e97de786-5683-796c-e32b-fcf4bbbf5ece","_type":"pet"},"from":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"},"to":{"_id":"8b951d6f-cd73-25f2-fb0a-cd5cb4d93937","_type":"dog","name":"charlie","weight":19}}
    dagger_test.go:157: {"node":{"_id":"5150f06e-b979-638e-5dd5-17a480001aee","_type":"friend"},"from":{"_id":"swash_1603572045463678000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"}}
    dagger_test.go:157: {"node":{"_id":"9e39c05a-3113-930a-4ff0-d841e382fbe4","_type":"fiance"},"from":{"_id":"cword_1603572045463636000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603572045463679000","_type":"user","name":"lacee"}}
```

## Roadmap
- [ ] import graph from JSON blob
- [ ] export graph to JSON blob