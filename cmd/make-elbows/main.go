// -*- compile-command: "go run main.go -stl ../../make-elbows.stl"; -*-

// make-elbows tests the STL output for several elbow-like structures.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	golden  = flag.Bool("golden", false, "Generate golden test files")
	objOut  = flag.String("obj", "extrude-helix.obj", "Output filename for Wavefront obj file")
	outBJK  = flag.String("o", "make-elbows.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut  = flag.String("stl", "make-elbows.stl", "Output filename for binary STL file")
	swapYZ  = flag.Bool("swapyz", false, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	if *golden {
		nodes.GenerateGoldenFilesPrefix = "golden-make-elbows"
	}

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
		AddNode("MakeBox.1", "origin=vector(0.5,1,0.5)", "size=vector(1,2,1)").
		AddNode("MakeBox.2", "origin=vector(1,0.5,0.5)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.1").
		Connect("MakeBox.1.out_mesh", "MergeMeshes.1.mesh_a").
		Connect("MakeBox.2.out_mesh", "MergeMeshes.1.mesh_b").
		//
		AddNode("MakeBox.3", "origin=vector(0.5,1,2.5)", "size=vector(1,2,1)").
		AddNode("MergeMeshes.2").
		Connect("MergeMeshes.1.out_mesh", "MergeMeshes.2.mesh_a").
		Connect("MakeBox.3.out_mesh", "MergeMeshes.2.mesh_b").
		//
		AddNode("MakeBox.4", "origin=vector(1.5,0.5,2.5)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.3").
		Connect("MergeMeshes.2.out_mesh", "MergeMeshes.3.mesh_a").
		Connect("MakeBox.4.out_mesh", "MergeMeshes.3.mesh_b").
		//
		AddNode("MakeBox.5", "origin=vector(0.5,1,4.5)", "size=vector(1,2,1)").
		AddNode("MergeMeshes.4").
		Connect("MergeMeshes.3.out_mesh", "MergeMeshes.4.mesh_a").
		Connect("MakeBox.5.out_mesh", "MergeMeshes.4.mesh_b").
		//
		AddNode("MakeBox.6", "origin=vector(2,0.5,4.5)", "size=vector(2,1,1)").
		AddNode("MergeMeshes.5").
		Connect("MergeMeshes.4.out_mesh", "MergeMeshes.5.mesh_a").
		Connect("MakeBox.6.out_mesh", "MergeMeshes.5.mesh_b").
		Build()
	must(err)

	if *outBJK == "-" {
		fmt.Printf("%v\n", design)
	} else if *outBJK != "" {
		must(os.WriteFile(*outBJK, []byte(design.String()+"\n"), 0644))
	}

	if *objOut != "" {
		must(c.ToObj(design, *objOut))
	}

	if *stlOut != "" {
		must(c.ToSTL(design, *stlOut, *swapYZ))
	}

	log.Printf("Done.")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
