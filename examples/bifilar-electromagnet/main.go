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
	innerDiam = flag.Float64("id", 3.0, "Inner diameter of first coil in millimeters")
	repoDir   = flag.String("repo", "/Users/glenn/src/github.com/gmlewis/blackjack", "Path to Blackjack repo")
	vertTurns = flag.Float64("vt", 2.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 0.5, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir)
	must(err)
	defer c.Close()

	log.Printf("Got %v nodes.", len(c.Nodes))

	design, err := c.NewBuilder().
		AddNode("Scalar.vert-turns", fmt.Sprintf("x=%v", *vertTurns)).
		AddNode("Point.helix-bbox", fmt.Sprintf("point=vector(%v,%v,%[1]v)", *innerDiam, 2**wireWidth)).
		AddNode("VectorMath.vert-gap", fmt.Sprintf("vec_b=vector(0,%v,0)", *wireGap)).
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap.point").
		AddNode("Helix.wire-1").
		Connect("VectorMath.vert-gap.out", "Helix.wire-1.size").
		AddNode("Helix.wire-2").
		AddNode("MakeQuad.wire-outline").
		AddNode("ExtrudeAlongCurve.wire-1").
		AddNode("ExtrudeAlongCurve.wire-1").
		Build()
	must(err)

	fmt.Printf("%v\n", design)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
