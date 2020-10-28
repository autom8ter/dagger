package dagger_test

import (
	"fmt"
	"github.com/autom8ter/dagger"
	"log"
)

func ExampleNewNode() {
	coleman = dagger.NewNode(map[string]interface{}{
		"_type": "user",
		"name":  "coleman",
	})
	tyler = dagger.NewNode(map[string]interface{}{
		"_type": "user",
		"name":  "coleman",
	})
	sarah = dagger.NewNode(map[string]interface{}{
		"_type": "user",
		"name":  "sarah",
	})
	lacee = dagger.NewNode(map[string]interface{}{
		"_type": "user",
		"name":  "lacee",
	})
	charlie = dagger.NewNode(map[string]interface{}{
		"_type":  "dog",
		"name":   "charlie",
		"weight": 25,
	})

	if _, err := coleman.Connect(tyler, "friend", true); err != nil {
		exitErr(err)
	}
	if _, err := sarah.Connect(lacee, "friend", true); err != nil {
		exitErr(err)
	}
	if _, err := coleman.Connect(lacee, "fiance", true); err != nil {
		exitErr(err)
	}
	if _, err := tyler.Connect(sarah, "wife", true); err != nil {
		exitErr(err)
	}
	if _, err := coleman.Connect(charlie, "pet", false); err != nil {
		exitErr(err)
	}
	if _, err := lacee.Connect(charlie, "pet", false); err != nil {
		exitErr(err)
	}
	if _, err := charlie.Connect(lacee, "owner", false); err != nil {
		exitErr(err)
	}
	if _, err := charlie.Connect(coleman, "owner", false); err != nil {
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
	fmt.Printf("registered node types = %v\n", dagger.NodeTypes())
	fmt.Printf("registered edge types = %v\n", dagger.EdgeTypes())

	// Output:
	// registered node types = [dog user]
	// registered edge types = [fiance friend owner pet wife]
}

func exit(msg string) {
	log.Fatal(msg)
}

func exitErr(err error) {
	log.Fatal(err)
}
