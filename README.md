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
	if err := tyler.Connect(lacee, "wife", true); err != nil {
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
registered node types = [user dog]registered edge types = [owner friend fiance wife pet]
found node = {"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}
{"node":{"_id":"63d853a5-306d-3fbd-360f-b35b9a8de632","_type":"friend"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603571323198158000","_type":"user","name":"sarah"}}
{"node":{"_id":"f6845cf5-52b1-508b-09c6-fff7588c1081","_type":"fiance"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"}}
{"node":{"_id":"b9f70c2f-b0e7-2f15-e07a-9f0bc2405092","_type":"wife"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"}}
{"node":{"_id":"d146c8f9-12b6-2b97-4342-61e41ba757f3","_type":"pet"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"ac3306e0-65bc-bc98-b2da-8fdbd7ce69e4","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"1a867925-d128-b622-6cd9-54b2f3feacbb","_type":"friend"},"from":{"_id":"swash_1603571323198158000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}
{"node":{"_id":"51a3b266-57d3-8095-d406-8ddd8818f0ae","_type":"fiance"},"from":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}
{"node":{"_id":"5ff29a8d-20f0-efdf-379a-e798f65ea473","_type":"wife"},"from":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}

found node = {"_id":"swash_1603571323197134000","_type":"user","name":"sarah"}
{"node":{"_id":"030b7df4-3898-0047-01ff-f976d8e2a6a8","_type":"friend"},"from":{"_id":"swash_1603571323197134000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}
{"node":{"_id":"869c044b-6198-0eb4-ab2c-98eda69de51e","_type":"friend"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603571323197134000","_type":"user","name":"sarah"}}

found node = {"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}
{"node":{"_id":"125944be-df7f-9f4d-12be-cd036ff212a3","_type":"wife"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"}}
{"node":{"_id":"6799c7ad-98dd-63b7-1edf-6111d3d4f0da","_type":"pet"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"558c8fa9-ca76-0616-781e-78bfd0cb8c74","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"869c044b-6198-0eb4-ab2c-98eda69de51e","_type":"friend"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603571323197134000","_type":"user","name":"sarah"}}
{"node":{"_id":"52a7758c-8161-926d-4e54-b5115c300d6a","_type":"fiance"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"}}
{"node":{"_id":"030b7df4-3898-0047-01ff-f976d8e2a6a8","_type":"friend"},"from":{"_id":"swash_1603571323197134000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}
{"node":{"_id":"149570e8-680d-2dcc-6235-c7d289a0933a","_type":"fiance"},"from":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}
{"node":{"_id":"8f709736-891d-1654-eb36-bd6c3b4c8915","_type":"wife"},"from":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}

found node = {"_id":"cword_1603571323197104000","_type":"user","name":"coleman"}
{"node":{"_id":"e0dcbe6c-591b-8c4c-b5b1-537771487ea4","_type":"friend"},"from":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"}}
{"node":{"_id":"149570e8-680d-2dcc-6235-c7d289a0933a","_type":"fiance"},"from":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}
{"node":{"_id":"c3981c89-2ad2-2671-b4a3-1cfe7b53b174","_type":"pet"},"from":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"},"to":{"_id":"558c8fa9-ca76-0616-781e-78bfd0cb8c74","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"fb96472e-1638-8e72-c7df-321de7b351ea","_type":"friend"},"from":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"}}
{"node":{"_id":"52a7758c-8161-926d-4e54-b5115c300d6a","_type":"fiance"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"}}

found node = {"_id":"twash_1603571323197132000","_type":"user","name":"tyler"}
{"node":{"_id":"8f709736-891d-1654-eb36-bd6c3b4c8915","_type":"wife"},"from":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"},"to":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"}}
{"node":{"_id":"fb96472e-1638-8e72-c7df-321de7b351ea","_type":"friend"},"from":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"}}
{"node":{"_id":"e0dcbe6c-591b-8c4c-b5b1-537771487ea4","_type":"friend"},"from":{"_id":"cword_1603571323197104000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"}}
{"node":{"_id":"125944be-df7f-9f4d-12be-cd036ff212a3","_type":"wife"},"from":{"_id":"ljans_1603571323197135000","_type":"user","name":"lacee"},"to":{"_id":"twash_1603571323197132000","_type":"user","name":"tyler"}}

found node = {"_id":"cword_1603571323198142000","_type":"user","name":"coleman"}
{"node":{"_id":"f2f168b2-26e3-4e9e-05f7-05c90559ee3d","_type":"friend"},"from":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"}}
{"node":{"_id":"51a3b266-57d3-8095-d406-8ddd8818f0ae","_type":"fiance"},"from":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}
{"node":{"_id":"50bc6a1a-01e9-5a4f-219d-1111bb56c2b9","_type":"pet"},"from":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"},"to":{"_id":"ac3306e0-65bc-bc98-b2da-8fdbd7ce69e4","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"551d3c9f-f323-3250-5ec9-1f9491ce3491","_type":"friend"},"from":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"}}
{"node":{"_id":"f6845cf5-52b1-508b-09c6-fff7588c1081","_type":"fiance"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"}}

found node = {"_id":"twash_1603571323198157000","_type":"user","name":"tyler"}
{"node":{"_id":"551d3c9f-f323-3250-5ec9-1f9491ce3491","_type":"friend"},"from":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"}}
{"node":{"_id":"5ff29a8d-20f0-efdf-379a-e798f65ea473","_type":"wife"},"from":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}
{"node":{"_id":"f2f168b2-26e3-4e9e-05f7-05c90559ee3d","_type":"friend"},"from":{"_id":"cword_1603571323198142000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"}}
{"node":{"_id":"b9f70c2f-b0e7-2f15-e07a-9f0bc2405092","_type":"wife"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"twash_1603571323198157000","_type":"user","name":"tyler"}}

found node = {"_id":"swash_1603571323198158000","_type":"user","name":"sarah"}
{"node":{"_id":"1a867925-d128-b622-6cd9-54b2f3feacbb","_type":"friend"},"from":{"_id":"swash_1603571323198158000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"}}
{"node":{"_id":"63d853a5-306d-3fbd-360f-b35b9a8de632","_type":"friend"},"from":{"_id":"ljans_1603571323198160000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603571323198158000","_type":"user","name":"sarah"}}
```

## Roadmap
- [ ] import graph from JSON blob
- [ ] export graph to JSON blob