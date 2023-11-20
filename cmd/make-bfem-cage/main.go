// -*- compile-command: "go run main.go -o '' -stl ../../make-bfem-cage.stl"; -*-

// make-bfem-cage tests the STL output for BFEMCage.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug    = flag.Bool("debug", false, "Turn on debugging info")
	golden   = flag.Bool("golden", false, "Generate golden test files")
	objOut   = flag.String("obj", "make-bfem-cage.obj", "Output filename for Wavefront obj file")
	outBJK   = flag.String("o", "make-bfem-cage.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir  = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	segments = flag.Int("ns", 36, "Number of segments")
	stlOut   = flag.String("stl", "make-bfem-cage.stl", "Output filename for binary STL file")
	swapYZ   = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	if *golden {
		nodes.GenerateGoldenFilesPrefix = "golden-make-bfem-cage"
	}

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	design, err := c.NewBuilder().AddNode("BFEMCage", set("segments", *segments)).Build()
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
