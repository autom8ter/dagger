# dagger [![GoDoc](https://godoc.org/github.com/autom8ter/dagger?status.svg)](https://godoc.org/github.com/autom8ter/dagger)

![dag](images/dag.png)

dagger is a blazing fast, concurrency safe, mutable, in-memory directed graph implementation with zero dependencies
    
    import "github.com/autom8ter/dagger"

## Design:

- flexibility
- global state
    - see [primitive](https://godoc.org/github.com/autom8ter/dagger/primitive) to manage graph state manually
- concurrency safe
- high performance
- simple api

## Features

- [x] native graph objects(nodes/edges)
- [x] typed graph objects(ex: user/pet)
- [x] labelled nodes & edges
- [x] depth first search
- [x] breadth first search
- [x] concurrency safe
- [x] import graph from JSON blob
- [x] export graph to JSON blob

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
   	coleman.EdgesFrom(dagger.AnyType(), func(e *dagger.Edge) bool {
   		if e.Type() == "pet" {
   			if e.To().GetInt("weight") != 19 {
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
   	coleman.EdgesFrom(dagger.AnyType(), func(e *dagger.Edge) bool {
   		if e.Type() == "pet" && e.GetString("name") == "charlie" {
   			exit("failed to delete node - (charlie)")
   		}
   		return true
   	})
   	// check to make sure edge no longer exists(cascade)
   	lacee.EdgesFrom(dagger.AnyType(), func(e *dagger.Edge) bool {
   		if e.Type() == "pet" && e.GetString("name") == "charlie" {
   			exit("failed to delete node - (charlie)")
   		}
   		return true
   	})

```

Output:
```

registered node types = [dog user]registered edge types = [fiance friend owner pet wife]
found node = {"_id":"cword_1603581855577975000","_type":"user","name":"coleman"}
{"node":{"_id":"aa91881a-1e23-9b44-e059-afe6880816a9","_type":"friend"},"from":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"}}
{"node":{"_id":"6ec23de7-9584-964b-79ed-ace7ee06112d","_type":"fiance"},"from":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"}}
{"node":{"_id":"7b86436f-28df-fff1-c198-8093397cc400","_type":"pet"},"from":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"},"to":{"_id":"85fb2e74-b125-243a-2ac8-8f13b1513824","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"1da7e6e1-197f-c65e-a6b9-335298b9c076","_type":"friend"},"from":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"}}
{"node":{"_id":"1bbfe7e0-6bf1-bca3-55e3-0281e80663b7","_type":"fiance"},"from":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"}}

found node = {"_id":"twash_1603581855577992000","_type":"user","name":"tyler"}
{"node":{"_id":"1da7e6e1-197f-c65e-a6b9-335298b9c076","_type":"friend"},"from":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"}}
{"node":{"_id":"2d6b7738-a549-5c8b-1a82-36e79ead4bc5","_type":"wife"},"from":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"}}
{"node":{"_id":"aa91881a-1e23-9b44-e059-afe6880816a9","_type":"friend"},"from":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"}}
{"node":{"_id":"3bfd44b4-3c73-0471-81be-1a09ff56e813","_type":"wife"},"from":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"}}

found node = {"_id":"cword_1603581855576640000","_type":"user","name":"coleman"}
{"node":{"_id":"001d526f-b4f0-6ac4-6d6f-063144400f3a","_type":"friend"},"from":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"}}
{"node":{"_id":"8c145e86-07c1-10f2-2062-6ba194277878","_type":"fiance"},"from":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"}}
{"node":{"_id":"b674a411-949d-6c6f-4279-985d46ab0b64","_type":"pet"},"from":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"},"to":{"_id":"60a2c505-fcc1-a649-66ac-df3378bba2ec","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"6d10e57f-bdee-7fe7-aa4d-ec1368775e09","_type":"friend"},"from":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"}}
{"node":{"_id":"3fbb68b7-a8f3-3313-b349-29596cf9a958","_type":"fiance"},"from":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"}}

found node = {"_id":"twash_1603581855578157000","_type":"user","name":"tyler"}
{"node":{"_id":"5371ab6e-8b8b-a2df-d12e-ecf852a083f7","_type":"wife"},"from":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"}}
{"node":{"_id":"87c893c3-2959-da05-a390-f0ca062ed90d","_type":"friend"},"from":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"}}
{"node":{"_id":"a793eeb7-4e35-5b06-b51f-916bf89d30ee","_type":"friend"},"from":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"}}
{"node":{"_id":"ab051d4e-b3a4-7d97-2d00-53b01c92bcbb","_type":"wife"},"from":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"}}

found node = {"_id":"swash_1603581855578158000","_type":"user","name":"sarah"}
{"node":{"_id":"e6671030-cb2a-80b8-5453-55991b87a1f6","_type":"friend"},"from":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"}}
{"node":{"_id":"ab051d4e-b3a4-7d97-2d00-53b01c92bcbb","_type":"wife"},"from":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"}}
{"node":{"_id":"5371ab6e-8b8b-a2df-d12e-ecf852a083f7","_type":"wife"},"from":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"}}
{"node":{"_id":"572ce11b-9c6e-06d0-876b-2cb43d4f9261","_type":"friend"},"from":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"}}

found node = {"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"}
{"node":{"_id":"99ed2178-e43c-ac5f-18be-a65855b35a6e","_type":"friend"},"from":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"}}
{"node":{"_id":"3fbb68b7-a8f3-3313-b349-29596cf9a958","_type":"fiance"},"from":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"}}
{"node":{"_id":"3ae8d958-2cc7-e9e5-fcbc-13b313418167","_type":"pet"},"from":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"},"to":{"_id":"60a2c505-fcc1-a649-66ac-df3378bba2ec","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"c0bb7318-a05a-1c96-4491-453645130c92","_type":"friend"},"from":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"}}
{"node":{"_id":"8c145e86-07c1-10f2-2062-6ba194277878","_type":"fiance"},"from":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"}}

found node = {"_id":"swash_1603581855577993000","_type":"user","name":"sarah"}
{"node":{"_id":"be911a98-106f-ca4b-1d6e-941715cda021","_type":"friend"},"from":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"}}
{"node":{"_id":"3bfd44b4-3c73-0471-81be-1a09ff56e813","_type":"wife"},"from":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"}}
{"node":{"_id":"4233d57b-3c32-d551-b904-90c5aee6f126","_type":"friend"},"from":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"}}
{"node":{"_id":"2d6b7738-a549-5c8b-1a82-36e79ead4bc5","_type":"wife"},"from":{"_id":"twash_1603581855577992000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"}}

found node = {"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"}
{"node":{"_id":"1bbfe7e0-6bf1-bca3-55e3-0281e80663b7","_type":"fiance"},"from":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"}}
{"node":{"_id":"bb244d0c-751b-9e71-5838-caa21d6c97ad","_type":"pet"},"from":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"},"to":{"_id":"85fb2e74-b125-243a-2ac8-8f13b1513824","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"4233d57b-3c32-d551-b904-90c5aee6f126","_type":"friend"},"from":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"}}
{"node":{"_id":"be911a98-106f-ca4b-1d6e-941715cda021","_type":"friend"},"from":{"_id":"swash_1603581855577993000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"}}
{"node":{"_id":"6ec23de7-9584-964b-79ed-ace7ee06112d","_type":"fiance"},"from":{"_id":"cword_1603581855577975000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855577994000","_type":"user","name":"lacee"}}

found node = {"_id":"twash_1603581855576673000","_type":"user","name":"tyler"}
{"node":{"_id":"39d8cf0b-ca55-da86-9579-9dc3e8dcb442","_type":"wife"},"from":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"}}
{"node":{"_id":"6d10e57f-bdee-7fe7-aa4d-ec1368775e09","_type":"friend"},"from":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"}}
{"node":{"_id":"001d526f-b4f0-6ac4-6d6f-063144400f3a","_type":"friend"},"from":{"_id":"cword_1603581855576640000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"}}
{"node":{"_id":"34b7bbd9-534c-31e8-219f-b53246ff0ed4","_type":"wife"},"from":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"}}

found node = {"_id":"swash_1603581855576675000","_type":"user","name":"sarah"}
{"node":{"_id":"34b7bbd9-534c-31e8-219f-b53246ff0ed4","_type":"wife"},"from":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"},"to":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"}}
{"node":{"_id":"c0bb7318-a05a-1c96-4491-453645130c92","_type":"friend"},"from":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"}}
{"node":{"_id":"39d8cf0b-ca55-da86-9579-9dc3e8dcb442","_type":"wife"},"from":{"_id":"twash_1603581855576673000","_type":"user","name":"tyler"},"to":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"}}
{"node":{"_id":"99ed2178-e43c-ac5f-18be-a65855b35a6e","_type":"friend"},"from":{"_id":"ljans_1603581855576676000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855576675000","_type":"user","name":"sarah"}}

found node = {"_id":"cword_1603581855578144000","_type":"user","name":"coleman"}
{"node":{"_id":"bfb8e908-01b7-2c5b-553b-15e2496bdd6d","_type":"pet"},"from":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"},"to":{"_id":"45d4829f-c351-ba1a-662a-289de972dcef","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"a793eeb7-4e35-5b06-b51f-916bf89d30ee","_type":"friend"},"from":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"},"to":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"}}
{"node":{"_id":"48300cc8-ac82-ab8f-7c00-4f3423e3f0b6","_type":"fiance"},"from":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"}}
{"node":{"_id":"87c893c3-2959-da05-a390-f0ca062ed90d","_type":"friend"},"from":{"_id":"twash_1603581855578157000","_type":"user","name":"tyler"},"to":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"}}
{"node":{"_id":"ffa5d8a0-e981-1836-7f83-a846c271dfe1","_type":"fiance"},"from":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"}}

found node = {"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"}
{"node":{"_id":"572ce11b-9c6e-06d0-876b-2cb43d4f9261","_type":"friend"},"from":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"},"to":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"}}
{"node":{"_id":"ffa5d8a0-e981-1836-7f83-a846c271dfe1","_type":"fiance"},"from":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"},"to":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"}}
{"node":{"_id":"7bf404ca-efe4-5d3e-81b8-297f9a4f3f2f","_type":"pet"},"from":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"},"to":{"_id":"45d4829f-c351-ba1a-662a-289de972dcef","_type":"dog","name":"charlie","weight":19}}
{"node":{"_id":"e6671030-cb2a-80b8-5453-55991b87a1f6","_type":"friend"},"from":{"_id":"swash_1603581855578158000","_type":"user","name":"sarah"},"to":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"}}
{"node":{"_id":"48300cc8-ac82-ab8f-7c00-4f3423e3f0b6","_type":"fiance"},"from":{"_id":"cword_1603581855578144000","_type":"user","name":"coleman"},"to":{"_id":"ljans_1603581855578159000","_type":"user","name":"lacee"}}

```
