// -*- compile-command: "go run main.go -o -"; -*-

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
	"os"
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug     = flag.Bool("debug", false, "Turn on debugging info")
	innerDiam = flag.Float64("id", 6.0, "Inner diameter of first coil in millimeters")
	numPairs  = flag.Int("np", 11, "Number of coil pairs (minimum 1)")
	numSegs   = flag.Int("ns", 36, "Number of segments per 360-degree turn of helix")
	outBJK    = flag.String("o", "bifilar-electromagnet.bjk", "Output filename for BJK file ('-' for stdout)")
	repoDir   = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut    = flag.String("stl", "", "Output filename for binary STL file")
	thickness = flag.Float64("th", 2.0, "Thickness of outer enclosing connecting wires in millimeters")
	vertTurns = flag.Float64("vt", 11.0, "Vertical turns of wire in electromagnet")
	wireGap   = flag.Float64("wg", 0.5, "Wire gap in millimeters")
	wireWidth = flag.Float64("ww", 1.0, "Wire width in millimeters")
)

func set(name string, value any) string {
	return fmt.Sprintf("%v=%v", name, value)
}

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
	if *outBJK == "" && *stlOut == "" {
		log.Fatalf("nothing to do: must supply -o or -stl or both")
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

	log.Printf("Got %v nodes.", len(c.Nodes))

	nodePosDY := 200
	nodePosY := -nodePosDY
	nextNodePos := func() string {
		nodePosY += nodePosDY
		return fmt.Sprintf("node_position=(0,%v)", nodePosY)
	}

	b := c.NewBuilder().
		// inputs that drive the rest of the design
		NewGroup("InnerRadius.inner-radius", embedNextNodePos(makeInnerRadius, nextNodePos)).
		//
		NewGroup("SizedQuad.wire-outline", embedNextNodePos(makeSizedQuad, nextNodePos)).
		//
		NewGroup("WireGaps.wire-gap", embedNextNodePos(makeWireGapNodes, nextNodePos)).
		//
		NewGroup("WireWidthAndGap.1", makeWireWidthAndGap).
		Connect("SizedQuad.wire-outline.wire-width-xz", "WireWidthAndGap.1.vec_a").
		Connect("WireGaps.wire-gap.vxz", "WireWidthAndGap.1.vec_b").
		//
		AddNode("MakeComment.vert-turns", nextNodePos(), "comment=This Scalar node controls\nthe number of vertical turns\nof the coil, thereby affecting\nits overall height.\nA value of 5\nseems to keep the UI pretty responsive.").
		AddNode("MakeScalar.vert-turns", set("x", *vertTurns)).
		//
		AddNode("MakeComment.thickness", nextNodePos(), "comment=This Scalar node controls\nthe thickness of the outer\nenclosing connecting wires\nin millimeters.").
		AddNode("MakeScalar.thickness", set("x", *thickness)).
		//
		AddNode("MakeComment.segments", nextNodePos(), "comment=This Scalar node controls\nthe number segments in a\nsingle turn of the coil.\nA value of 36\nseems to keep the UI pretty responsive.").
		AddNode("MakeScalar.segments", set("x", *numSegs)).
		//
		AddNode("VectorMath.inner-rad-half-ww").
		Connect("InnerRadius.inner-radius.vxz", "VectorMath.inner-rad-half-ww.vec_a").
		Connect("SizedQuad.wire-outline.half-wire-width-xz", "VectorMath.inner-rad-half-ww.vec_b").
		//
		AddNode("VectorMath.helix-bbox").
		Connect("VectorMath.inner-rad-half-ww.out", "VectorMath.helix-bbox.vec_a").
		Connect("SizedQuad.wire-outline.twice-wire-width-y", "VectorMath.helix-bbox.vec_b").
		//
		AddNode("VectorMath.vert-gap").
		Connect("VectorMath.helix-bbox.out", "VectorMath.vert-gap.vec_a").
		Connect("WireGaps.wire-gap.twice_vy", "VectorMath.vert-gap.vec_b").
		// define a pair of coils
		AddNode("MakeComment", nextNodePos(), "comment=This is coil pair #1:").
		NewGroup("CoilPair.coils-1-2", makeCoilPair, "delta_y=0", "start_angle=0").
		// external controlling connections
		Connect("MakeScalar.vert-turns.x", "CoilPair.coils-1-2.turns").
		Connect("MakeScalar.segments.x", "CoilPair.coils-1-2.segments").
		Connect("VectorMath.vert-gap.out", "CoilPair.coils-1-2.size").
		Connect("SizedQuad.wire-outline.out_mesh", "CoilPair.coils-1-2.cross_section").
		Connect("SizedQuad.wire-outline.wire-width", "CoilPair.coils-1-2.wire_width").
		Connect("WireGaps.wire-gap.wire_gap", "CoilPair.coils-1-2.wire_gap").
		MergeMesh("CoilPair.coils-1-2.out_mesh")

	lastSizeOut := "VectorMath.vert-gap.out"
	nodePosDY = 600
	for i := 2; i <= *numPairs; i++ {
		pairName := fmt.Sprintf("pair-%v", i)
		sizeMathNode := fmt.Sprintf("VectorMath.size-%v", pairName)
		thisCoilPair := fmt.Sprintf("CoilPair.%v", pairName)
		b = b.
			// second instance of CoilPair
			AddNode("MakeComment", nextNodePos(), fmt.Sprintf("comment=This is coil pair #%v:", i)).
			AddNode(sizeMathNode).
			Connect("WireWidthAndGap.1.vxz", sizeMathNode+".vec_b").
			Connect(lastSizeOut, sizeMathNode+".vec_a").
			AddNode(thisCoilPair, set("delta_y", float64(i-1)/float64(*numPairs-1)), set("start_angle", 180.0*float64(i-1)/float64(*numPairs))).
			Connect("MakeScalar.vert-turns.x", thisCoilPair+".turns").
			Connect("MakeScalar.segments.x", thisCoilPair+".segments").
			Connect(sizeMathNode+".out", thisCoilPair+".size").
			Connect("SizedQuad.wire-outline.out_mesh", thisCoilPair+".cross_section").
			Connect("SizedQuad.wire-outline.wire-width", thisCoilPair+".wire_width").
			Connect("WireGaps.wire-gap.wire_gap", thisCoilPair+".wire_gap").
			MergeMesh(thisCoilPair + ".out_mesh")
		lastSizeOut = sizeMathNode + ".out"
	}

	b = b.
		AddNode("BFEMCage.cage", set("num_pairs", *numPairs)).
		Connect(lastSizeOut, "BFEMCage.cage.size").
		Connect("SizedQuad.wire-outline.wire-width", "BFEMCage.cage.wire_width").
		Connect("WireGaps.wire-gap.wire_gap", "BFEMCage.cage.wire_gap").
		Connect("MakeScalar.segments.x", "BFEMCage.cage.segments").
		Connect("MakeScalar.vert-turns.x", "BFEMCage.cage.turns").
		Connect("MakeScalar.thickness.x", "BFEMCage.cage.thickness").
		MergeMesh("BFEMCage.cage.out_mesh")

	design, err := b.Build()
	must(err)

	if *outBJK == "-" {
		fmt.Printf("%v\n", design)
	} else if *outBJK != "" {
		must(os.WriteFile(*outBJK, []byte(design.String()+"\n"), 0644))
	}

	if *stlOut != "" {
		buf, err := c.ToSTL(design)
		must(err)
		must(os.WriteFile(*stlOut, buf, 0644))
	}

	log.Printf("Done.")
}

func makeCoilPair(b *nodes.Builder) *nodes.Builder {
	return b.
		AddNode("ScalarMath.add180", "op=Add", "y=180").
		AddNode("ScalarMath.add-width-gap", "op=Add").
		AddNode("ScalarMath.mul-delta-y", "op=Mul").
		AddNode("MakeVector.pos-y").
		AddNode("Helix.wire-1").
		AddNode("Helix.wire-2").
		AddNode("ExtrudeAlongCurve.wire-1", "flip=1").
		AddNode("ExtrudeAlongCurve.wire-2", "flip=1").
		AddNode("MergeMeshes.wire-1-2").
		// internal connections
		Connect("ScalarMath.add180.out", "Helix.wire-1.start_angle").
		Connect("ScalarMath.add-width-gap.out", "ScalarMath.mul-delta-y.x").
		Connect("ScalarMath.mul-delta-y.out", "MakeVector.pos-y.y").
		Connect("MakeVector.pos-y.v", "Helix.wire-1.pos").
		Connect("MakeVector.pos-y.v", "Helix.wire-2.pos").
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
		Input("delta_y", "ScalarMath.mul-delta-y.y").
		Input("wire_width", "ScalarMath.add-width-gap.x").
		Input("wire_gap", "ScalarMath.add-width-gap.y").
		Output("MergeMeshes.wire-1-2.out_mesh", "out_mesh")
}

func embedNextNodePos(f embedNNPFunc, nextNodePos func() string) nodes.BuilderFunc {
	return func(b *nodes.Builder) *nodes.Builder { return f(b, nextNodePos) }
}

type embedNNPFunc func(b *nodes.Builder, nextNodePos func() string) *nodes.Builder

func makeSizedQuad(b *nodes.Builder, nextNodePos func() string) *nodes.Builder {
	return b.
		AddNode("MakeComment.wire-width", nextNodePos(), "comment=This Scalar node controls the\nwidth of the wire in mm.").
		AddNode("MakeScalar.wire-width", set("x", *wireWidth)).
		AddNode("MakeVector.wire-width").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.x").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.y").
		Connect("MakeScalar.wire-width.x", "MakeVector.wire-width.z").
		AddNode("MakeQuad.wire-outline", "normal=vector(0,0,1)").
		Connect("MakeVector.wire-width.v", "MakeQuad.wire-outline.size").
		Output("MakeQuad.wire-outline.out_mesh", "out_mesh").
		AddNode("VectorMath.mul-half-xz", "op=Mul", "vec_b=vector(0.5,0,0.5)").
		AddNode("VectorMath.mul-1-xz", "op=Mul", "vec_b=vector(1,0,1)").
		AddNode("VectorMath.mul-2-y", "op=Mul", "vec_b=vector(0,2,0)").
		Connect("MakeVector.wire-width.v", "VectorMath.mul-half-xz.vec_a").
		Connect("MakeVector.wire-width.v", "VectorMath.mul-1-xz.vec_a").
		Connect("MakeVector.wire-width.v", "VectorMath.mul-2-y.vec_a").
		Output("MakeScalar.wire-width.x", "wire-width").
		Output("VectorMath.mul-half-xz.out", "half-wire-width-xz").
		Output("VectorMath.mul-1-xz.out", "wire-width-xz").
		Output("VectorMath.mul-2-y.out", "twice-wire-width-y")
}

func makeWireGapNodes(b *nodes.Builder, nextNodePos func() string) *nodes.Builder {
	return b.
		AddNode("MakeComment.wire-gap", nextNodePos(), "comment=This Scalar node controls\nthe gap between wires\nin millimeters.").
		AddNode("MakeScalar.wire-gap", set("x", *wireGap)).
		AddNode("MakeVector.wire-gap-y").
		AddNode("MakeVector.wire-gap-xz").
		Connect("MakeScalar.wire-gap.x", "MakeVector.wire-gap-y.y").
		Connect("MakeScalar.wire-gap.x", "MakeVector.wire-gap-xz.x").
		Connect("MakeScalar.wire-gap.x", "MakeVector.wire-gap-xz.z").
		AddNode("VectorMath.twice-vy", "op=Mul", "vec_b=vector(0,2,0)").
		Connect("MakeVector.wire-gap-y.v", "VectorMath.twice-vy.vec_a").
		Output("MakeScalar.wire-gap.x", "wire_gap").
		Output("VectorMath.twice-vy.out", "twice_vy").
		Output("MakeVector.wire-gap-xz.v", "vxz")
}

func makeWireWidthAndGap(b *nodes.Builder) *nodes.Builder {
	return b.
		AddNode("VectorMath", "op=Add").
		Input("vec_a", "VectorMath.vec_a").
		Input("vec_b", "VectorMath.vec_b").
		Output("VectorMath.out", "vxz")
}

func makeInnerRadius(b *nodes.Builder, nextNodePos func() string) *nodes.Builder {
	innerRadius := 0.5 * *innerDiam
	return b.
		AddNode("MakeComment.inner-radius", nextNodePos(), "comment=This Scalar node controls\nthe radius of the inner-most\ncoil pair in millimeters.").
		AddNode("MakeScalar.inner-radius", set("x", innerRadius)).
		AddNode("MakeVector.inner-radius-xz").
		Connect("MakeScalar.inner-radius.x", "MakeVector.inner-radius-xz.x").
		Connect("MakeScalar.inner-radius.x", "MakeVector.inner-radius-xz.z").
		Output("MakeVector.inner-radius-xz.v", "vxz")
}

var _ embedNNPFunc = makeSizedQuad

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
