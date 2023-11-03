// -*- compile-command: "go run main.go -debug -stl ../../extrude-helix.stl"; -*-

// extrude-helix tests the STL output for MakeQuad + Helix + extrude
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
	stlOut  = flag.String("stl", "extrude-helix.stl", "Output filename for binary STL file")
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
		AddNode("MakeQuad.1", "normal=vector(0,0,1)").
		AddNode("Helix.1", "segments=8").
		AddNode("ExtrudeAlongCurve.1", "flip=1").
		Connect("MakeQuad.1.out_mesh", "ExtrudeAlongCurve.1.cross_section").
		Connect("Helix.1.out_mesh", "ExtrudeAlongCurve.1.backbone").
		Build()
	must(err)

	fmt.Printf("%v\n", design)

	if *stlOut != "" {
		must(c.ToSTL(design, *stlOut))
	}

	log.Printf("Done.")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
