// -*- compile-command: "go run main.go"; -*-

// make-scalar tests the Eval output for MakeScalar.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	value   = flag.Float64("val", 33, "Value to set scalar to")
)

func main() {
	flag.Parse()

	homeDir, err := homedir.Dir()
	must(err)

	repoPath := filepath.Join(homeDir, *repoDir)
	c, err := nodes.New(repoPath, *debug)
	if err != nil {
		c, err = nodes.New(*repoDir, *debug)
		must(err)
	}
	defer c.Close()

	design, err := c.NewBuilder().AddNode("MakeScalar", fmt.Sprintf("x=%v", *value)).Build()
	must(err)
	mesh, err := c.Eval(design)
	must(err)
	log.Printf("Mesh=%#v", mesh)

	log.Printf("Done.")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
