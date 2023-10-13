// -*- compile-command: "go run main.go"; -*-

// bifilar-electromagnet generates a Blackjack .bjk file that
// represents a bifilar electromagnet similar to those found
// here: https://github.com/gmlewis/irmf-examples/tree/master/examples/012-bifilar-electromagnet#axial-radial-bifilar-electromagnet-1irmf
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/gmlewis/go-bjk/nodes"
)

var (
	repoDir = flag.String("repo", "/Users/glenn/src/github.com/gmlewis/blackjack", "Path to Blackjack repo")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ns, err := c.List()
	must(err)
	log.Printf("Got %v nodes.", len(ns))

	design := ast.New()
	fmt.Printf("%v\n", design)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
