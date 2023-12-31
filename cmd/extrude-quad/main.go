// -*- compile-command: "go run main.go -stl ../../extrude-quad.stl"; -*-

// extrude-quad tests the STL output for MakeQuad + extrude
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	amount  = flag.Float64("mm", 1, "Millimeters to extrude quad")
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	golden  = flag.Bool("golden", false, "Generate golden test files")
	objOut  = flag.String("obj", "extrude-quad.obj", "Output filename for Wavefront obj file")
	outBJK  = flag.String("o", "extrude-quad.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut  = flag.String("stl", "extrude-quad.stl", "Output filename for binary STL file")
	swapYZ  = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	if *golden {
		nodes.GenerateGoldenFilesPrefix = "golden-extrude-quad"
	}

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	design, err := c.NewBuilder().
		AddNode("MakeQuad.1").
		AddNode("ExtrudeFacesWithCaps.1", set("amount", *amount), "faces=*").
		Connect("MakeQuad.1.out_mesh", "ExtrudeFacesWithCaps.1.in_mesh").
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

func set(key string, value any) string {
	return fmt.Sprintf("%v=%v", key, value)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
