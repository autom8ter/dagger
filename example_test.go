package dagger_test

import (
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/primitive"
	"log"
	"time"
)

func ExampleNewNode() {
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

	if err := coleman.Connect(tyler, "friend"); err != nil {
		exitErr(err)
	}
	if err := tyler.Connect(coleman, "friend"); err != nil {
		exitErr(err)
	}
	if err := sarah.Connect(lacee, "friend"); err != nil {
		exitErr(err)
	}
	if err := lacee.Connect(sarah, "friend"); err != nil {
		exitErr(err)
	}
	if err := coleman.Connect(lacee, "fiance"); err != nil {
		exitErr(err)
	}
	if err := lacee.Connect(coleman, "fiance"); err != nil {
		exitErr(err)
	}
	if err := tyler.Connect(lacee, "wife"); err != nil {
		exitErr(err)
	}
	if err := sarah.Connect(tyler, "wife"); err != nil {
		exitErr(err)
	}
	if err := coleman.Connect(charlie, "pet"); err != nil {
		exitErr(err)
	}
	if err := lacee.Connect(charlie, "pet"); err != nil {
		exitErr(err)
	}
	if err := charlie.Connect(lacee, "owner"); err != nil {
		exitErr(err)
	}
	if err := charlie.Connect(coleman, "owner"); err != nil {
		exitErr(err)
	}
	charlie.Patch(map[string]interface{}{
		"weight": 19,
	})
	// check to make sure edge is patched
	coleman.EdgesFrom(func(e *primitive.Edge) bool {
		if e.Type() == "pet" && e.GetString("name") == "charlie" {
			if e.To.GetInt("weight") != 19 {
				exit("failed to patch charlie's weight")
			}
		}
		return true
	})
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
}

func exit(msg string) {
	log.Fatal(msg)
}

func exitErr(err error) {
	log.Fatal(err)
}
