package nodes

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/google/go-cmp/cmp"
)

const (
	repoPath = "src/github.com/gmlewis/blackjack"
)

var (
	c *Client
)

func TestMain(m *testing.M) {
	var err error
	c, err = New(repoPath, false)
	if err != nil {
		log.Fatalf("unable to create test Client: %v", err)
	}
	defer c.Close()

	os.Exit(m.Run())
}

//go:embed testdata/bifilar-electromagnet.bjk
var bifilarElectromagnet string

func TestBuild(t *testing.T) {
	t.Parallel()
	if c == nil {
		t.Fatalf("c is nil")
	}
	design, err := c.NewBuilder().
		// nodes:
		AddNode("MakeQuad.wire-outline", "size=vector(1,1,1)", "normal=vector(0,0,1)").   // node_idx: 0
		AddNode("Helix.wire-1", "start_angle=180", "segments=36", "direction=Clockwise"). // node_idx: 1
		AddNode("Helix.wire-2", "start_angle=180", "segments=36").                        // node_idx: 2
		AddNode("ExtrudeAlongCurve.wire-2", "flip=1").                                    // node_idx: 3
		AddNode("Helix.wire-3", "start_angle=0", "segments=36").                          // node_idx: 4
		AddNode("ExtrudeAlongCurve.wire-3", "flip=1").                                    // node_idx: 5
		AddNode("MergeMeshes.wire-2-3").                                                  // node_idx: 6
		AddNode("MakeScalar.vert-turns", "x=2").                                          // node_idx: 7
		AddNode("ExtrudeAlongCurve.wire-1", "flip=1").                                    // node_idx: 8
		AddNode("MergeMeshes.wire-23-1").                                                 // node_idx: 9
		AddNode("Point.helix-bbox", "point=vector(3,2,3)").                               // node_idx: 10
		AddNode("VectorMath.vert-gap-1", "vec_b=vector(0,0.5,0)").                        // node_idx: 11
		AddNode("VectorMath.vert-gap-2", "vec_b=vector(1.5,0,1.5)").                      // node_idx: 12
		AddNode("Helix.wire-4", "start_angle=0", "segments=36").                          // node_idx: 13
		AddNode("ExtrudeAlongCurve.wire-4", "flip=1").                                    // node_idx: 14
		AddNode("MergeMeshes.wire-231-4").                                                // node_idx: 15
		// connections:
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap-1.vec_a").
		Connect("VectorMath.vert-gap-1.out", "VectorMath.vert-gap-2.vec_a").
		Connect("VectorMath.vert-gap-2.out", "Helix.wire-1.size").
		Connect("VectorMath.vert-gap-1.out", "Helix.wire-2.size").
		Connect("VectorMath.vert-gap-1.out", "Helix.wire-3.size").
		Connect("VectorMath.vert-gap-2.out", "Helix.wire-4.size").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-1.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-2.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-3.turns").
		Connect("MakeScalar.vert-turns.x", "Helix.wire-4.turns").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-1.cross_section").
		Connect("Helix.wire-1.out_mesh", "ExtrudeAlongCurve.wire-1.backbone").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-2.cross_section").
		Connect("Helix.wire-2.out_mesh", "ExtrudeAlongCurve.wire-2.backbone").
		Connect("ExtrudeAlongCurve.wire-2.out_mesh", "MergeMeshes.wire-2-3.mesh_a").
		Connect("ExtrudeAlongCurve.wire-3.out_mesh", "MergeMeshes.wire-2-3.mesh_b").
		Connect("Helix.wire-3.out_mesh", "ExtrudeAlongCurve.wire-3.backbone").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-3.cross_section").
		Connect("Helix.wire-4.out_mesh", "ExtrudeAlongCurve.wire-4.backbone").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-4.cross_section").
		Connect("MergeMeshes.wire-2-3.out_mesh", "MergeMeshes.wire-23-1.mesh_a").
		Connect("ExtrudeAlongCurve.wire-1.out_mesh", "MergeMeshes.wire-23-1.mesh_b").
		Connect("MergeMeshes.wire-23-1.out_mesh", "MergeMeshes.wire-231-4.mesh_a").
		Connect("ExtrudeAlongCurve.wire-4.out_mesh", "MergeMeshes.wire-231-4.mesh_b").
		Build()
	if err != nil {
		t.Fatal(err)
	}

	// Force the UIData to match the manually-generated test data.
	ui := design.Graph.UIData
	ui.NodePositions = []*ast.Vec2{
		{X: 1072.3687, Y: 232.1065},
		{X: 1070.6208, Y: 990.6514},
		{X: 1075.0774, Y: 478.61154},
		{X: 1625.3439, Y: 371.60907},
		{X: 1073.4645, Y: 734.81415},
		{X: 1618.6683, Y: 734.70935},
		{X: 1941.8724, Y: 571.3904},
		{X: 689.5807, Y: 813.1497},
		{X: 1620.0642, Y: 1022.223},
		{X: 2180.4314, Y: 699.56116},
		{X: -261.1836, Y: 290.5234},
		{X: 83.387024, Y: 306.02115},
		{X: 558.4893, Y: 1026.147},
		{X: 1076.6501, Y: 1242.5769},
		{X: 1618.2654, Y: 1269.7031},
		{X: 2386.4956, Y: 821.07043},
	}
	ui.NodeOrder = []uint64{2, 3, 4, 5, 0, 1, 8, 9, 6, 7, 11, 10, 12, 13, 14, 15}
	ui.Pan = ast.Vec2{X: 914.03564, Y: -222.5001}
	ui.Zoom = 1.9877489

	keyedParamValues := map[string]*ast.ParamValue{}
	for _, pv := range design.Graph.ExternalParameters.ParamValues {
		key := fmt.Sprintf("%v,%v", pv.NodeIdx, pv.ParamName)
		keyedParamValues[key] = pv
	}
	var sortedParamValues []*ast.ParamValue
	for _, key := range wantParamsSortOrder {
		sortedParamValues = append(sortedParamValues, keyedParamValues[key])
	}
	design.Graph.ExternalParameters.ParamValues = sortedParamValues

	got, want := design.String(), bifilarElectromagnet
	if diff := cmp.Diff(want, got); diff != "" {
		t.Log("\n\n" + got + "\n")
		t.Errorf("design mismatch (-want +got):\n%v", diff)
	}
}

var wantParamsSortOrder = []string{
	"0,right",
	"0,center",
	"5,flip",
	"13,start_angle",
	"2,direction",
	"1,start_angle",
	"1,direction",
	"12,vec_b",
	"4,segments",
	"13,direction",
	"0,size",
	"10,point",
	"13,segments",
	"4,start_angle",
	"13,pos",
	"4,direction",
	"1,segments",
	"12,op",
	"2,pos",
	"14,flip",
	"2,start_angle",
	"1,pos",
	"0,normal",
	"2,segments",
	"11,op",
	"7,x",
	"4,pos",
	"3,flip",
	"8,flip",
	"11,vec_b",
}
