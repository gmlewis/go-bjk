// -*- compile-command: "go run main.go -stl ../../extrude-helix.stl > ../../extrude-helix.bjk"; -*-

// extrude-helix tests the STL output for MakeQuad + Helix + extrude
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug     = flag.Bool("debug", false, "Turn on debugging info")
	golden    = flag.Bool("golden", false, "Generate golden test files")
	innerDiam = flag.Float64("id", 6.0, "Inner diameter of first coil in millimeters")
	numSegs   = flag.Int("ns", 36, "Number of segments per 360-degree turn of helix")
	objOut    = flag.String("obj", "extrude-helix.obj", "Output filename for Wavefront obj file")
	outBJK    = flag.String("o", "extrude-helix.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir   = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut    = flag.String("stl", "extrude-helix.stl", "Output filename for binary STL file")
	swapYZ    = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
	vertTurns = flag.Float64("vt", 1.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 5.0, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func main() {
	flag.Parse()

	if *golden {
		nodes.GenerateGoldenFilesPrefix = "golden-extrude-helix"
	}

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	innerRadius := 0.5 * *innerDiam
	sx := innerRadius + 0.5**wireWidth
	sy := 2 * (*wireWidth + *wireGap)
	sz := sx

	design, err := c.NewBuilder().
		AddNode("MakeQuad.1", "normal=vector(0,0,1)").
		AddNode("Helix.1", setv("size", sx, sy, sz), set("segments", *numSegs), set("turns", *vertTurns)).
		AddNode("Helix.2", setv("size", sx, sy, sz), set("segments", *numSegs), set("turns", *vertTurns), "start_angle=180").
		AddNode("ExtrudeAlongCurve.1", "flip=1").
		AddNode("ExtrudeAlongCurve.2", "flip=1").
		Connect("MakeQuad.1.out_mesh", "ExtrudeAlongCurve.1.cross_section").
		Connect("Helix.1.out_mesh", "ExtrudeAlongCurve.1.backbone").
		Connect("MakeQuad.1.out_mesh", "ExtrudeAlongCurve.2.cross_section").
		Connect("Helix.2.out_mesh", "ExtrudeAlongCurve.2.backbone").
		AddNode("MergeMeshes.1").
		Connect("ExtrudeAlongCurve.1.out_mesh", "MergeMeshes.1.mesh_a").
		Connect("ExtrudeAlongCurve.2.out_mesh", "MergeMeshes.1.mesh_b").
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

func setv(key string, x, y, z any) string {
	return fmt.Sprintf("%v=vector(%v,%v,%v)", key, x, y, z)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
