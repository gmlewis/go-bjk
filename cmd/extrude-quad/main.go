// -*- compile-command: "go run main.go -stl ../../extrude-quad.stl"; -*-

// extrude-quad tests the STL output for MakeQuad + extrude
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
	amount  = flag.Float64("mm", 1, "Millimeters to extrude quad")
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut  = flag.String("stl", "extrude-quad.stl", "Output filename for binary STL file")
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

	design, err := c.NewBuilder().
		AddNode("MakeQuad.1").
		AddNode("ExtrudeFacesWithCaps.1", set("amount", *amount), "faces=*").
		Connect("MakeQuad.1.out_mesh", "ExtrudeFacesWithCaps.1.in_mesh").
		Build()
	must(err)

	fmt.Printf("%v\n", design)

	if *stlOut != "" {
		must(c.ToSTL(design, *stlOut))
	}

	log.Printf("Done.")
}

func set(key string, value any) string {
	return fmt.Sprintf("%v=%v", key, value)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}