package nodes

import (
	_ "embed"
	"testing"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/google/go-cmp/cmp"
)

const (
	// TODO: make this on a machine other than my laptop?!?
	repoDir = "/Users/glenn/src/github.com/gmlewis/blackjack"
)

//go:embed testdata/bifilar-electromagnet.bjk
var bifilarElectromagnet string

func TestBuild(t *testing.T) {
	c, err := New(repoDir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	design, err := c.NewBuilder().
		AddNode("MakeQuad.wire-outline", "size=vector(1,1,1)").
		AddNode("Helix.wire-1", "start_angle=180", "segments=36").
		AddNode("Helix.wire-2", "start_angle=0", "segments=36").
		AddNode("ExtrudeAlongCurve.wire-1").
		AddNode("Helix.wire-3", "start_angle=180", "segments=36").
		AddNode("ExtrudeAlongCurve.wire-2").
		AddNode("MergeMeshes.wire-1-2").
		AddNode("MakeScalar.vert-turns", "x=2").
		AddNode("ExtrudeAlongCurve.wire-3").
		AddNode("MergeMeshes.wire-2-3").
		AddNode("Point.helix-bbox", "point=vector(3,2,3)").
		AddNode("VectorMath.vert-gap-1", "vec_b=vector(0,0.5,0)").
		AddNode("VectorMath.vert-gap-2", "vec_b=vector(0,1.5,0)").
		AddNode("Helix.wire-4", "start_angle=0", "segments=36").
		AddNode("ExtrudeAlongCurve.wire-4").
		AddNode("MergeMeshes.wire-3-4").
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap-1.vec_a").
		Connect("VectorMath.vert-gap-1.out", "Helix.wire-1.size").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-1.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-2.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-3.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-4.turns").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-1.cross_section").
		Connect("Helix.wire-1.out_mesh", "ExtrudeAlongCurve.wire-1.backbone").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-2.cross_section").
		Connect("Helix.wire-2.out_mesh", "ExtrudeAlongCurve.wire-2.backbone").
		Connect("ExtrudeAlongCurve.wire-1.out_mesh", "MergeMeshes.wire-1-2.mesh_a").
		Connect("ExtrudeAlongCurve.wire-2.out_mesh", "MergeMeshes.wire-1-2.mesh_b").
		Build()
	if err != nil {
		t.Fatal(err)
	}

	// Force the UIData to match the manually-generated test data.
	ui := design.Graph.UIData
	ui.NodePositions = []*ast.Vec2{
		{1072.3687, 232.1065},
		{1070.6208, 990.6514},
		{1075.0774, 478.61154},
		{1625.3439, 371.60907},
		{1073.4645, 734.81415},
		{1618.6683, 734.70935},
		{1941.8724, 571.3904},
		{689.5807, 813.1497},
		{1620.0642, 1022.223},
		{2180.4314, 699.56116},
		{-261.1836, 290.5234},
		{83.387024, 306.02115},
		{558.4893, 1026.147},
		{1076.6501, 1242.5769},
		{1618.2654, 1269.7031},
		{2386.4956, 821.07043},
	}
	ui.NodeOrder = []uint64{2, 3, 4, 5, 0, 1, 8, 9, 6, 7, 11, 10, 12, 13, 14, 15}
	ui.Pan = ast.Vec2{914.03564, -222.5001}
	ui.Zoom = 1.9877489

	got, want := design.String(), bifilarElectromagnet
	if diff := cmp.Diff(want, got); diff != "" {
		t.Log("\n\n" + got + "\n")
		t.Errorf("design mismatch (-want +got):\n%v", diff)
	}
}
