// -*- compile-command: "go run main.go -stl ../../make-elbows.stl"; -*-

// make-elbows tests the STL output for several elbow-like structures.
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
	stlOut  = flag.String("stl", "make-elbows.stl", "Output filename for binary STL file")
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
		AddNode("MakeBox.1", "origin=vector(0,0.5,0)", "size=vector(1,2,1)").
		AddNode("MakeBox.2", "origin=vector(0.5,0,0)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.1").
		Connect("MakeBox.1.out_mesh", "MergeMeshes.1.mesh_a").
		Connect("MakeBox.2.out_mesh", "MergeMeshes.1.mesh_b").
		//
		AddNode("MakeBox.3", "origin=vector(0,0.5,2)", "size=vector(1,2,1)").
		AddNode("MergeMeshes.2").
		Connect("MergeMeshes.1.out_mesh", "MergeMeshes.2.mesh_a").
		Connect("MakeBox.3.out_mesh", "MergeMeshes.2.mesh_b").
		//
		AddNode("MakeBox.4", "origin=vector(1,0,2)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.3").
		Connect("MergeMeshes.2.out_mesh", "MergeMeshes.3.mesh_a").
		Connect("MakeBox.4.out_mesh", "MergeMeshes.3.mesh_b").
		//
		AddNode("MakeBox.5", "origin=vector(0,0.5,4)", "size=vector(1,2,1)").
		AddNode("MergeMeshes.4").
		Connect("MergeMeshes.3.out_mesh", "MergeMeshes.4.mesh_a").
		Connect("MakeBox.5.out_mesh", "MergeMeshes.4.mesh_b").
		//
		AddNode("MakeBox.6", "origin=vector(1.5,0,4)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.5").
		Connect("MergeMeshes.4.out_mesh", "MergeMeshes.5.mesh_a").
		Connect("MakeBox.6.out_mesh", "MergeMeshes.5.mesh_b").
		Build()
	must(err)

	// fmt.Printf("%v\n", design)

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
