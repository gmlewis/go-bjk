package nodes

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/google/go-cmp/cmp"
)

const (
	// TODO: make this work on a machine other than my laptop?!?
	repoDir = "/Users/glenn/src/github.com/gmlewis/blackjack"
)

var (
	c *Client
)

func TestMain(m *testing.M) {
	var err error
	c, err = New(repoDir, true)
	if err != nil {
		log.Fatal("unable to create test Client")
	}
	defer c.Close()

	os.Exit(m.Run())
}

func TestAddNode(t *testing.T) {
	t.Parallel()
	inputsByType := map[string]string{}
	outputsByType := map[string]string{}

	// First, generate a map of all possible BJK input/output types along
	// with an input and an output of that type, then try every possible
	// combination to make sure that go-bjk enforces type checking during "Connect".
	for _, node := range c.Nodes {
		for _, input := range node.Inputs {
			if _, ok := inputsByType[input.DataType]; ok {
				continue
			}
			inputsByType[input.DataType] = fmt.Sprintf("%v.%v", node.OpName, input.Name)
		}
		for _, output := range node.Outputs {
			if _, ok := outputsByType[output.DataType]; ok {
				continue
			}
			if in, ok := inputsByType[output.DataType]; ok && strings.HasPrefix(in, node.OpName) {
				continue
			}
			outputsByType[output.DataType] = fmt.Sprintf("%v.%v", node.OpName, output.Name)
		}
	}

	t.Logf("Found %v input types: %+v", len(inputsByType), inputsByType)
	t.Logf("Found %v output types: %+v", len(outputsByType), outputsByType)

	shouldErr := func(outNode, outType, inType string) bool {
		out, ok := c.Nodes[outNode]
		if !ok {
			t.Fatalf("programming error: could not find node %q", outNode)
		}
		// If any of the output node's input pins is of type "mesh", then this pin
		// will be unconnected and will cause an error.
		// Since we iterate over maps to choose the inNodes above, this outcome
		// can be different on each test.
		for _, input := range out.Inputs {
			if input.DataType == "mesh" {
				return true
			}
		}

		return outType != inType
		// TODO: Maybe lua_string can connect to enum or selection or string inputs?
	}

	for outType, nodeOutPin := range outputsByType {
		parts := strings.Split(nodeOutPin, ".")
		outNode := parts[0]
		for inType, nodeInPin := range inputsByType {
			parts = strings.Split(nodeInPin, ".")
			inNode := parts[0]

			name := fmt.Sprintf("Connect('%v','%v')", nodeOutPin, nodeInPin)
			t.Run(name, func(t *testing.T) {
				wantErr := shouldErr(outNode, outType, inType)
				_, err := c.NewBuilder().
					AddNode(outNode).
					AddNode(inNode).
					Connect(nodeOutPin, nodeInPin).
					Build()
				if wantErr && err == nil {
					t.Fatal("Connect worked, want error")
				}
				if !wantErr && err != nil {
					t.Fatalf("Connect = %v, want success", err)
				}
			})
		}
	}
}

//go:embed testdata/bifilar-electromagnet.bjk
var bifilarElectromagnet string

func TestBuild(t *testing.T) {
	t.Parallel()
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
