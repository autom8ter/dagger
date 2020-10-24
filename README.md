# dagger [![GoDoc](https://godoc.org/github.com/autom8ter/dagger?status.svg)](https://godoc.org/github.com/autom8ter/dagger)

dagger is a blazing fast, concurrency safe, mutable, in-memory directed graph implementation with zero dependencies
    
    import "github.com/autom8ter/dagger"


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

```