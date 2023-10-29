// -*- compile-command: "go run main.go"; -*-

// make-box tests the STL output for MakeBox.
package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
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

	design, err := c.NewBuilder().AddNode("MakeBox").Build()
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
