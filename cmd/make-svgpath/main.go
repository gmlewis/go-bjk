// -*- compile-command: "go run main.go -o '' -stl '' -obj ../../make-svgpath.obj"; -*-

// make-svgpath tests the STL output for SVGPath.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

// test sword from: https://yqnn.github.io/svg-path-editor/
// M 4 8 L 10 1 L 13 0 L 12 3 L 5 9 C 6 10 6 11 7 10 C 7 11 8 12 7 12 A 1.42 1.42 0 0 1 6 13 A 5 5 0 0 0 4 10 Q 3.5 9.9 3.5 10.5 T 2 11.8 T 1.2 11 T 2.5 9.5 T 3 9 A 5 5 90 0 0 0 7 A 1.42 1.42 0 0 1 1 6 C 1 5 2 6 3 6 C 2 7 3 7 4 8 M 10 1 L 10 3 L 12 3 L 10.2 2.8 L 10 1
// test examples from: https://www.w3.org/TR/SVG2/paths.html
// M300,200 h-150 a150,150 0 1,0 150,-150 z
// M275,175 v-150 a150,150 0 0,0 -150,150 z

var (
	dPath = flag.String("d",
		"M 4 8 L 10 1 L 13 0 L 12 3 L 5 9 C 6 10 6 11 7 10 C 7 11 8 12 7 12 A 1.42 1.42 0 0 1 6 13 A 5 5 0 0 0 4 10 Q 3.5 9.9 3.5 10.5 T 2 11.8 T 1.2 11 T 2.5 9.5 T 3 9 A 5 5 90 0 0 0 7 A 1.42 1.42 0 0 1 1 6 C 1 5 2 6 3 6 C 2 7 3 7 4 8 M 10 1 L 10 3 L 12 3 L 10.2 2.8 L 10 1",
		"SVG path ('d') to generate")
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	golden  = flag.Bool("golden", false, "Generate golden test files")
	objOut  = flag.String("obj", "make-svgpath.obj", "Output filename for Wavefront obj file")
	outBJK  = flag.String("o", "make-svgpath.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut  = flag.String("stl", "make-svgpath.stl", "Output filename for binary STL file")
	swapYZ  = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	if *dPath == "" {
		log.Fatal("empty SVG -d path - nothing to output")
	}

	if *golden {
		nodes.GenerateGoldenFilesPrefix = "golden-make-svgpath"
	}

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	design, err := c.NewBuilder().AddNode("SVGPath", set("d", *dPath)).Build()
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
