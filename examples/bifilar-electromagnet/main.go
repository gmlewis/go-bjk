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
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug     = flag.Bool("debug", false, "Turn on debugging info")
	innerDiam = flag.Float64("id", 6.0, "Inner diameter of first coil in millimeters")
	numPairs  = flag.Int("np", 11, "Number of coil pairs (minimum 1)")
	numSegs   = flag.Int("ns", 36, "Number of segments per 360-degree turn of helix")
	repoDir   = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo")
	vertTurns = flag.Float64("vt", 5.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 0.5, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func main() {
	flag.Parse()

	if *innerDiam < 0 {
		log.Fatalf("-id must be at least 0")
	}
	if *numPairs < 1 {
		log.Fatalf("-np must be at least 1")
	}
	if *numSegs < 1 {
		log.Fatalf("-ns must be at least 1")
	}
	if *vertTurns < 0 {
		log.Fatalf("-vt must be at least 0")
	}

	homeDir, err := homedir.Dir()
	must(err)

	repoPath := filepath.Join(homeDir, *repoDir)
	c, err := nodes.New(repoPath, *debug)
	must(err)
	defer c.Close()

	log.Printf("Got %v nodes.", len(c.Nodes))

	var nodePosY int
	nextNodePos := func() string {
		s := fmt.Sprintf("node_position=(0,%v)", nodePosY)
		nodePosY += 200
		return s
	}

	innerRadius := 0.5 * *innerDiam
	b := c.NewBuilder().
		// inputs that drive the rest of the design
		NewGroup("SizedQuad.wire-outline", embedNextNodePos(makeSizedQuad, nextNodePos)).
		AddNode("MakeComment.vert-turns", nextNodePos(), "comment=This Scalar node controls\nthe number of vertical turns\nof the coil, thereby affecting\nits overall height.\nA value of 5\nseems to keep the UI pretty responsive.").
		AddNode("MakeScalar.vert-turns", fmt.Sprintf("x=%v", *vertTurns)).
		AddNode("MakeComment.segments", nextNodePos(), "comment=This Scalar node controls\nthe number segments in a\nsingle turn of the coil.\nA value of 36\nseems to keep the UI pretty responsive.").
		AddNode("MakeScalar.segments", fmt.Sprintf("x=%v", *numSegs)).
		AddNode("MakeComment.start-angle-shift-mixer", nextNodePos(), "comment=This Scalar node controls\nthe mix from\nno rotation (0) of successive\ncoils to max (1) rotation.").
		AddNode("MakeScalar.start-angle-shift-mixer", "x=1", "min=-1", "max=1").
		AddNode("Point.helix-bbox", fmt.Sprintf("point=vector(%v,%v,%[1]v)", innerRadius+0.5**wireWidth, 2**wireWidth)).
		AddNode("VectorMath.vert-gap", fmt.Sprintf("vec_b=vector(0,%v,0)", *wireGap)).
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap.vec_a").
		// define a pair of coils
		AddNode("MakeComment", nextNodePos(), "comment=This is coil pair #1:").
		NewGroup("CoilPair.coils-1-2", makeCoilPair).
		// external controlling connections
		Connect("MakeScalar.vert-turns.x", "CoilPair.coils-1-2.turns").
		Connect("MakeScalar.segments.x", "CoilPair.coils-1-2.segments").
		Connect("VectorMath.vert-gap.out", "CoilPair.coils-1-2.size").
		Connect("SizedQuad.wire-outline.out_mesh", "CoilPair.coils-1-2.cross_section")

	lastMergeMeshes := "CoilPair.coils-1-2"
	for i := 2; i <= *numPairs; i++ {
		pairName := fmt.Sprintf("pair-%v", i)
		sizeMathNode := fmt.Sprintf("VectorMath.size-%v", pairName)
		coilStartAngleMixerNode := fmt.Sprintf("ScalarMath.start-angle-mixer-%v", pairName)
		thisCoilPair := fmt.Sprintf("CoilPair.%v", pairName)
		thisMergeMeshes := fmt.Sprintf("MergeMeshes.%v", pairName)
		b = b.
			// second instance of CoilPair
			AddNode("MakeComment", fmt.Sprintf("node_position=(0,%v)", 600*i), fmt.Sprintf("comment=This is coil pair #%v:", i)).
			AddNode(coilStartAngleMixerNode, "op=Mul", fmt.Sprintf("y=%v", 180.0*float64(i-1)/float64(*numPairs))).
			AddNode(sizeMathNode, fmt.Sprintf("vec_b=vector(%v,0,%[1]v)", float64(i-1)*(*wireWidth+*wireGap))).
			Connect("VectorMath.vert-gap.out", sizeMathNode+".vec_a").
			AddNode(thisCoilPair).
			Connect("MakeScalar.vert-turns.x", thisCoilPair+".turns").
			Connect("MakeScalar.segments.x", thisCoilPair+".segments").
			Connect("MakeScalar.start-angle-shift-mixer.x", coilStartAngleMixerNode+".x").
			Connect(coilStartAngleMixerNode+".out", thisCoilPair+".start_angle").
			Connect(sizeMathNode+".out", thisCoilPair+".size").
			Connect("SizedQuad.wire-outline.out_mesh", thisCoilPair+".cross_section").
			AddNode(thisMergeMeshes).
			Connect(lastMergeMeshes+".out_mesh", thisMergeMeshes+".mesh_a").
			Connect(thisCoilPair+".out_mesh", thisMergeMeshes+".mesh_b")
		lastMergeMeshes = thisMergeMeshes
	}

	design, err := b.Build()
	must(err)

	fmt.Printf("%v\n", design)
}

func makeCoilPair(b *nodes.Builder) *nodes.Builder {
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
}

func embedNextNodePos(f embedNNPFunc, nextNodePos func() string) nodes.BuilderFunc {
	return func(b *nodes.Builder) *nodes.Builder { return f(b, nextNodePos) }
}

type embedNNPFunc func(b *nodes.Builder, nextNodePos func() string) *nodes.Builder

func makeSizedQuad(b *nodes.Builder, nextNodePos func() string) *nodes.Builder {
	return b.
		AddNode("MakeComment.wire-width", nextNodePos(), "comment=This Scalar node controls the\nwidth of the wire in mm.").
		AddNode("MakeScalar.wire-width", fmt.Sprintf("x=%v", *wireWidth)).
		AddNode("MakeVector.wire-width").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.x").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.y").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.z").
		AddNode("MakeQuad.wire-outline", "normal=vector(0,0,1)").
		Connect("MakeVector.wire-width.v", "MakeQuad.wire-outline.size").
		Output("MakeQuad.wire-outline.out_mesh", "out_mesh")
}

var _ embedNNPFunc = makeSizedQuad

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
