// -*- compile-command: "go run main.go"; -*-

// bifilar-electromagnet generates a Blackjack .bjk file that
// represents a bifilar electromagnet similar to those found
// here: https://github.com/gmlewis/irmf-examples/tree/master/examples/012-bifilar-electromagnet#axial-radial-bifilar-electromagnet-1irmf
//
// Warning - increasing the numPairs, numSegs, or vertTurns parameters
// can cause significant degredation in the performance of Blackjack
// when viewing the designs.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug     = flag.Bool("debug", false, "Turn on debugging info")
	innerDiam = flag.Float64("id", 6.0, "Inner diameter of first coil in millimeters")
	numPairs  = flag.Int("np", 11, "Number of coil pairs (minimum 1)")
	numSegs   = flag.Int("ns", 36, "Number of segments per 360-degree turn of helix")
	repoDir   = flag.String("repo", "/Users/glenn/src/github.com/gmlewis/blackjack", "Path to Blackjack repo")
	vertTurns = flag.Float64("vt", 5.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 0.5, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	log.Printf("Got %v nodes.", len(c.Nodes))

	innerRadius := 0.5 * *innerDiam
	b := c.NewBuilder().
		// inputs that drive the rest of the design
		AddNode("MakeQuad.wire-outline", fmt.Sprintf("size=vector(%v,%[1]v,%[1]v)", *wireWidth), "normal=vector(0,0,1)").
		AddNode("MakeScalar.vert-turns", fmt.Sprintf("x=%v", *vertTurns)).
		AddNode("MakeScalar.segments", fmt.Sprintf("x=%v", *numSegs)).
		AddNode("MakeScalar.start-angle", "x=0").
		AddNode("MakeScalar.start-angle-shift-mixer", "x=1", "min=-1", "max=1").
		AddNode("Point.helix-bbox", fmt.Sprintf("point=vector(%v,%v,%[1]v)", innerRadius+0.5**wireWidth, 2**wireWidth)).
		AddNode("VectorMath.vert-gap", fmt.Sprintf("vec_b=vector(0,%v,0)", *wireGap)).
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap.vec_a").
		// define a pair of coils
		NewGroup("CoilPair", func(b *nodes.Builder) *nodes.Builder {
			return b.
				AddNode("ScalarMath.add180", "op=Add", "y=180").
				AddNode("Helix.wire-1").
				AddNode("Helix.wire-2").
				AddNode("ExtrudeAlongCurve.wire-1", "flip=1").
				AddNode("ExtrudeAlongCurve.wire-2", "flip=1").
				AddNode("MergeMeshes.wire-1-2").
				// internal connections
				Connect("ScalarMath.add180.out", "Helix.wire-1.start_angle").
				Connect("Helix.wire-1.out_mesh", "ExtrudeAlongCurve.wire-1.backbone").
				Connect("Helix.wire-2.out_mesh", "ExtrudeAlongCurve.wire-2.backbone").
				Connect("ExtrudeAlongCurve.wire-1.out_mesh", "MergeMeshes.wire-1-2.mesh_a").
				Connect("ExtrudeAlongCurve.wire-2.out_mesh", "MergeMeshes.wire-1-2.mesh_b").
				Input("cross_section", "ExtrudeAlongCurve.wire-1.cross_section").
				Input("cross_section", "ExtrudeAlongCurve.wire-2.cross_section").
				Input("turns", "Helix.wire-1.turns").
				Input("turns", "Helix.wire-2.turns").
				Input("size", "Helix.wire-1.size").
				Input("size", "Helix.wire-2.size").
				Input("segments", "Helix.wire-1.segments").
				Input("segments", "Helix.wire-2.segments").
				Input("start_angle", "ScalarMath.add180.x").
				Input("start_angle", "Helix.wire-2.start_angle").
				Output("MergeMeshes.wire-1-2.out_mesh", "out_mesh")
		}).
		// instance of group
		AddNode("CoilPair.coils-1-2").
		// external controlling connections
		Connect("MakeScalar.vert-turns.x", "CoilPair.coils-1-2.turns").
		Connect("MakeScalar.segments.x", "CoilPair.coils-1-2.segments").
		Connect("MakeScalar.start-angle.x", "CoilPair.coils-1-2.start_angle").
		Connect("VectorMath.vert-gap.out", "CoilPair.coils-1-2.size").
		Connect("MakeQuad.wire-outline.out_mesh", "CoilPair.coils-1-2.cross_section")

	lastMergeMeshes := "CoilPair.coils-1-2"
	for i := 2; i < *numPairs; i++ {
		pairName := fmt.Sprintf("pair-%v", i)
		sizeMathNode := fmt.Sprintf("VectorMath.size-%v", pairName)
		// coilStartAngleNode := fmt.Sprintf("MakeScalar.start-angle-%v", pairName)
		coilStartAngleMixerNode := fmt.Sprintf("ScalarMath.start-angle-mixer-%v", pairName)
		thisCoilPair := fmt.Sprintf("CoilPair.%v", pairName)
		thisMergeMeshes := fmt.Sprintf("MergeMeshes.%v", pairName)
		b = b.
			// second instance of CoilPair
			AddNode(sizeMathNode, fmt.Sprintf("vec_b=vector(%v,0,%[1]v)", float64(i-1)*(*wireWidth+*wireGap))).
			Connect("VectorMath.vert-gap.out", sizeMathNode+".vec_a").
			AddNode(thisCoilPair).
			AddNode(coilStartAngleMixerNode, "op=Mul", fmt.Sprintf("y=%v", 180.0*float64(i-1)/float64(*numPairs-1))).
			// AddNode(coilStartAngleNode).
			Connect("MakeScalar.vert-turns.x", thisCoilPair+".turns").
			Connect("MakeScalar.segments.x", thisCoilPair+".segments").
			Connect("MakeScalar.start-angle-shift-mixer.x", coilStartAngleMixerNode+".x").
			Connect(coilStartAngleMixerNode+".out", thisCoilPair+".start_angle").
			Connect(sizeMathNode+".out", thisCoilPair+".size").
			Connect("MakeQuad.wire-outline.out_mesh", thisCoilPair+".cross_section").
			AddNode(thisMergeMeshes).
			Connect(lastMergeMeshes+".out_mesh", thisMergeMeshes+".mesh_a").
			Connect(thisCoilPair+".out_mesh", thisMergeMeshes+".mesh_b")
		lastMergeMeshes = thisMergeMeshes
	}

	design, err := b.Build()
	must(err)

	fmt.Printf("%v\n", design)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
