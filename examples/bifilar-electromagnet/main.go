// -*- compile-command: "go run main.go"; -*-

// bifilar-electromagnet generates a Blackjack .bjk file that
// represents a bifilar electromagnet similar to those found
// here: https://github.com/gmlewis/irmf-examples/tree/master/examples/012-bifilar-electromagnet#axial-radial-bifilar-electromagnet-1irmf
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug     = flag.Bool("debug", false, "Turn on debugging info")
	innerDiam = flag.Float64("id", 3.0, "Inner diameter of first coil in millimeters")
	numSegs   = flag.Int("ns", 36, "Number of segments per 360-degree turn of helix")
	repoDir   = flag.String("repo", "/Users/glenn/src/github.com/gmlewis/blackjack", "Path to Blackjack repo")
	vertTurns = flag.Float64("vt", 2.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 0.5, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	log.Printf("Got %v nodes.", len(c.Nodes))

	design, err := c.NewBuilder().
		// nodes
		AddNode("MakeQuad.wire-outline", fmt.Sprintf("size=vector(%v,%[1]v,%[1]v)", *wireWidth), "normal=vector(0,0,1)").
		AddNode("Helix.wire-1", "start_angle=180", fmt.Sprintf("segments=%v", *numSegs)).
		AddNode("Helix.wire-2", "start_angle=0", fmt.Sprintf("segments=%v", *numSegs)).
		AddNode("ExtrudeAlongCurve.wire-1", "flip=1").
		AddNode("ExtrudeAlongCurve.wire-2", "flip=1").
		AddNode("MergeMeshes.wire-1-2").
		// connections
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-1.cross_section").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-2.cross_section").
		Connect("Helix.wire-1.out_mesh", "ExtrudeAlongCurve.wire-1.backbone").
		Connect("Helix.wire-2.out_mesh", "ExtrudeAlongCurve.wire-2.backbone").
		Connect("ExtrudeAlongCurve.wire-1.out_mesh", "MergeMeshes.wire-1-2.mesh_a").
		Connect("ExtrudeAlongCurve.wire-2.out_mesh", "MergeMeshes.wire-1-2.mesh_b").
		// AddNode("MakeScalar.vert-turns", fmt.Sprintf("x=%v", *vertTurns)).
		// AddNode("Point.helix-bbox", fmt.Sprintf("point=vector(%v,%v,%[1]v)", *innerDiam, 2**wireWidth)).
		// AddNode("VectorMath.vert-gap", fmt.Sprintf("vec_b=vector(0,%v,0)", *wireGap)).
		// Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-1.cross_section").
		// Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-2.cross_section").
		// Connect("Point.helix-bbox.point", "VectorMath.vert-gap.vec_a").
		// Connect("VectorMath.vert-gap.out", "Helix.wire-1.size").
		Build()
	must(err)

	fmt.Printf("%v\n", design)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
