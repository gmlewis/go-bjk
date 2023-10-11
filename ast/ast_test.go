package ast

import (
	_ "embed"
	"os"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/hexops/valast"
)

//go:embed "testdata/bifilar-electromagnet.bjk"
var testFile string

const header = `// BLACKJACK_VERSION_HEADER 0 0 0
`

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *BJK
	}{
		{
			name:  "no graph",
			input: header + "( )",
			want:  &BJK{},
		},
		{
			name:  "no nodes",
			input: header + "( nodes: [ ] )",
			want:  &BJK{Graph: &Graph{}},
		},
		{
			name:  "one node no return value",
			input: header + `( nodes: [ ( op_name: "MakeQuad", return_value: None, inputs: [ ], outputs: [ ], ) ] )`,
			want: &BJK{Graph: &Graph{
				Nodes: []*Node{{OpName: "MakeQuad"}},
			}},
		},
		{
			name:  "one node no inputs no outputs",
			input: header + `( nodes: [ ( op_name: "MakeQuad", return_value: Some("out_mesh"), inputs: [ ], outputs: [ ], ) ] )`,
			want: &BJK{Graph: &Graph{
				Nodes: []*Node{{OpName: "MakeQuad", ReturnValue: String("out_mesh")}},
			}},
		},
		{
			name:  "one node one input",
			input: header + `( nodes: [ ( op_name: "MakeQuad", return_value: None, inputs: [ ( name: "center", data_type: "BJK_VECTOR", kind: External( promoted: None, ), ), ], outputs: [ ], ) ] )`,
			want: &BJK{
				Graph: &Graph{
					Nodes: []*Node{{
						OpName: "MakeQuad",
						Inputs: []*Input{{
							Name:     "center",
							DataType: "BJK_VECTOR",
							Kind:     DependencyKind{External: &External{}},
						}},
					}},
				},
			},
		},
		{
			name:  "one node one output",
			input: header + `( nodes: [ ( op_name: "MakeQuad", return_value: None, inputs: [ ], outputs: [ ( name: "out_mesh", data_type: "BJK_MESH", ), ], ), ], )`,
			want: &BJK{
				Graph: &Graph{
					Nodes: []*Node{{
						OpName: "MakeQuad",
						Outputs: []*Output{{
							Name:     "out_mesh",
							DataType: "BJK_MESH",
						}},
					}},
				},
			},
		},
		{
			name:  "one node one input with kind connection",
			input: header + `( nodes: [ ( op_name: "MakeQuad", return_value: None, inputs: [ ( name: "center", data_type: "BJK_VECTOR", kind: Conection( node_idx: 3, param_name: "out_mesh", ), ), ], outputs: [ ], ) ] )`,
			want: &BJK{
				Graph: &Graph{
					Nodes: []*Node{{
						OpName: "MakeQuad",
						Inputs: []*Input{{
							Name:     "center",
							DataType: "BJK_VECTOR",
							Kind: DependencyKind{
								Connection: &Connection{
									NodeIdx:   3,
									ParamName: "out_mesh",
								},
							},
						}},
					}},
				},
			},
		},
		{
			name:  "default node",
			input: header + "( nodes: [ ], default_node: Some(15), )",
			want:  &BJK{Graph: &Graph{DefaultNode: Uint64(15)}},
		},
		{
			name:  "no default node",
			input: header + "( nodes: [ ], default_node: None, )",
			want:  &BJK{Graph: &Graph{}},
		},
		{
			name:  "ui_data",
			input: header + "( nodes: [ ], default_node: Some(15), ui_data: Some(( node_positions: [ (1072.3687, 232.1065), (1070.6208, 990.6514), ], node_order: [ 2, 3, ], pan: (-0.6252365, -565.1915), zoom: 1.342995, locked_gizmo_nodes: [], )), )",
			want: &BJK{
				Graph: &Graph{
					DefaultNode: Uint64(15),
					UIData: &UIData{
						NodePositions: []*Vec2{{X: 1072.3687, Y: 232.1065}, {X: 1070.6208, Y: 990.6514}},
						NodeOrder:     []uint64{2, 3},
						Pan:           Vec2{X: -0.6252365, Y: -565.1915},
						Zoom:          1.342995,
					},
				},
			},
		},
		{
			name:  "complete testFile example",
			input: testFile,
			want:  testFileWant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parser.ParseString("", tt.input, participle.Trace(os.Stderr))
			if err != nil {
				t.Logf("%v\n", tt.input)
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Logf("%v\n", tt.input)
				t.Logf("\n%v\n", valast.String(got))
				t.Errorf("ParseString mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func String(s string) *string { return &s }
func Uint64(v uint64) *uint64 { return &v }

var testFileWant = &BJK{
	Version: Version{
		Minor: 1,
	},
	Graph: &Graph{
		Nodes: []*Node{
			{
				OpName:      "MakeQuad",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "center",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "normal",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "right",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "size",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "Helix",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "pos",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "size",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   12,
							ParamName: "out",
						}},
					},
					{
						Name:     "start_angle",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "turns",
						DataType: "BJK_SCALAR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   7,
							ParamName: "x",
						}},
					},
					{
						Name:     "segments",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "direction",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "Helix",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "pos",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "size",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   11,
							ParamName: "out",
						}},
					},
					{
						Name:     "start_angle",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "turns",
						DataType: "BJK_SCALAR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   7,
							ParamName: "x",
						}},
					},
					{
						Name:     "segments",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "direction",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "ExtrudeAlongCurve",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "backbone",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   2,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "cross_section",
						DataType: "BJK_MESH",
						Kind:     DependencyKind{Connection: &Connection{ParamName: "out_mesh"}},
					},
					{
						Name:     "flip",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "Helix",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "pos",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "size",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   11,
							ParamName: "out",
						}},
					},
					{
						Name:     "start_angle",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "turns",
						DataType: "BJK_SCALAR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   7,
							ParamName: "x",
						}},
					},
					{
						Name:     "segments",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "direction",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "ExtrudeAlongCurve",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "backbone",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   4,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "cross_section",
						DataType: "BJK_MESH",
						Kind:     DependencyKind{Connection: &Connection{ParamName: "out_mesh"}},
					},
					{
						Name:     "flip",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "MergeMeshes",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "mesh_a",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   3,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "mesh_b",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   5,
							ParamName: "out_mesh",
						}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName: "MakeScalar",
				Inputs: []*Input{{
					Name:     "x",
					DataType: "BJK_SCALAR",
					Kind:     DependencyKind{External: &External{}},
				}},
				Outputs: []*Output{{
					Name:     "x",
					DataType: "BJK_SCALAR",
				}},
			},
			{
				OpName:      "ExtrudeAlongCurve",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "backbone",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   1,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "cross_section",
						DataType: "BJK_MESH",
						Kind:     DependencyKind{Connection: &Connection{ParamName: "out_mesh"}},
					},
					{
						Name:     "flip",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "MergeMeshes",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "mesh_a",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   6,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "mesh_b",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   8,
							ParamName: "out_mesh",
						}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName: "Point",
				Inputs: []*Input{{
					Name:     "point",
					DataType: "BJK_VECTOR",
					Kind:     DependencyKind{External: &External{}},
				}},
				Outputs: []*Output{{
					Name:     "point",
					DataType: "BJK_VECTOR",
				}},
			},
			{
				OpName: "VectorMath",
				Inputs: []*Input{
					{
						Name:     "op",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "vec_a",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   10,
							ParamName: "point",
						}},
					},
					{
						Name:     "vec_b",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out",
					DataType: "BJK_VECTOR",
				}},
			},
			{
				OpName: "VectorMath",
				Inputs: []*Input{
					{
						Name:     "op",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "vec_a",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   11,
							ParamName: "out",
						}},
					},
					{
						Name:     "vec_b",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out",
					DataType: "BJK_VECTOR",
				}},
			},
			{
				OpName:      "Helix",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "pos",
						DataType: "BJK_VECTOR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "size",
						DataType: "BJK_VECTOR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   12,
							ParamName: "out",
						}},
					},
					{
						Name:     "start_angle",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "turns",
						DataType: "BJK_SCALAR",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   7,
							ParamName: "x",
						}},
					},
					{
						Name:     "segments",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
					{
						Name:     "direction",
						DataType: "BJK_STRING",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "ExtrudeAlongCurve",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "backbone",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   13,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "cross_section",
						DataType: "BJK_MESH",
						Kind:     DependencyKind{Connection: &Connection{ParamName: "out_mesh"}},
					},
					{
						Name:     "flip",
						DataType: "BJK_SCALAR",
						Kind:     DependencyKind{External: &External{}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
			{
				OpName:      "MergeMeshes",
				ReturnValue: String("out_mesh"),
				Inputs: []*Input{
					{
						Name:     "mesh_a",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   9,
							ParamName: "out_mesh",
						}},
					},
					{
						Name:     "mesh_b",
						DataType: "BJK_MESH",
						Kind: DependencyKind{Connection: &Connection{
							NodeIdx:   14,
							ParamName: "out_mesh",
						}},
					},
				},
				Outputs: []*Output{{
					Name:     "out_mesh",
					DataType: "BJK_MESH",
				}},
			},
		},
		DefaultNode: Uint64(15),
		UIData: &UIData{
			NodePositions: []*Vec2{
				{
					X: 1072.3687,
					Y: 232.1065,
				},
				{
					X: 1070.6208,
					Y: 990.6514,
				},
				{
					X: 1075.0774,
					Y: 478.61154,
				},
				{
					X: 1625.3439,
					Y: 371.60907,
				},
				{
					X: 1073.4645,
					Y: 734.81415,
				},
				{
					X: 1618.6683,
					Y: 734.70935,
				},
				{
					X: 1941.8724,
					Y: 571.3904,
				},
				{
					X: 689.5807,
					Y: 813.1497,
				},
				{
					X: 1620.0642,
					Y: 1022.223,
				},
				{
					X: 2180.4314,
					Y: 699.56116,
				},
				{
					X: -261.1836,
					Y: 290.5234,
				},
				{
					X: 83.387024,
					Y: 306.02115,
				},
				{
					X: 558.4893,
					Y: 1026.147,
				},
				{
					X: 1076.6501,
					Y: 1242.5769,
				},
				{
					X: 1618.2654,
					Y: 1269.7031,
				},
				{
					X: 2386.4956,
					Y: 821.07043,
				},
			},
			NodeOrder: []uint64{
				2,
				3,
				4,
				5,
				0,
				1,
				8,
				9,
				6,
				7,
				11,
				10,
				12,
				13,
				14,
				15,
			},
			Pan: Vec2{
				X: -0.6252365,
				Y: -565.1915,
			},
			Zoom: 1.342995,
		},
		ExternalParameters: &ExternalParameters{ParamValues: []*ParamValue{
			{
				NodeIdx:   10,
				ParamName: "point",
				ValueEnum: ValueEnum{Vector: &VectorValue{
					X: 3,
					Y: 2,
					Z: 3,
				}},
			},
			{
				ParamName: "center",
				ValueEnum: ValueEnum{Vector: &VectorValue{}},
			},
			{
				NodeIdx:   4,
				ParamName: "segments",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 36}},
			},
			{
				ParamName: "size",
				ValueEnum: ValueEnum{Vector: &VectorValue{
					X: 1,
					Y: 1,
					Z: 1,
				}},
			},
			{
				NodeIdx:   1,
				ParamName: "direction",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Clockwise"}},
			},
			{
				NodeIdx:   14,
				ParamName: "flip",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 1}},
			},
			{
				NodeIdx:   4,
				ParamName: "pos",
				ValueEnum: ValueEnum{Vector: &VectorValue{}},
			},
			{
				NodeIdx:   11,
				ParamName: "op",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Add"}},
			},
			{
				NodeIdx:   4,
				ParamName: "direction",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Clockwise"}},
			},
			{
				NodeIdx:   13,
				ParamName: "direction",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Clockwise"}},
			},
			{
				NodeIdx:   2,
				ParamName: "segments",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 36}},
			},
			{
				NodeIdx:   13,
				ParamName: "pos",
				ValueEnum: ValueEnum{Vector: &VectorValue{}},
			},
			{
				NodeIdx:   1,
				ParamName: "segments",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 36}},
			},
			{
				NodeIdx:   5,
				ParamName: "flip",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 1}},
			},
			{
				NodeIdx:   4,
				ParamName: "start_angle",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{}},
			},
			{
				NodeIdx:   7,
				ParamName: "x",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 2}},
			},
			{
				NodeIdx:   12,
				ParamName: "op",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Add"}},
			},
			{
				NodeIdx:   13,
				ParamName: "segments",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 36}},
			},
			{
				NodeIdx:   11,
				ParamName: "vec_b",
				ValueEnum: ValueEnum{Vector: &VectorValue{Y: 0.5}},
			},
			{
				ParamName: "right",
				ValueEnum: ValueEnum{Vector: &VectorValue{X: 1}},
			},
			{
				NodeIdx:   2,
				ParamName: "pos",
				ValueEnum: ValueEnum{Vector: &VectorValue{}},
			},
			{
				NodeIdx:   13,
				ParamName: "start_angle",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{}},
			},
			{
				NodeIdx:   2,
				ParamName: "direction",
				ValueEnum: ValueEnum{Selection: &SelectionValue{Selection: "Clockwise"}},
			},
			{
				NodeIdx:   8,
				ParamName: "flip",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 1}},
			},
			{
				NodeIdx:   3,
				ParamName: "flip",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 1}},
			},
			{
				NodeIdx:   1,
				ParamName: "start_angle",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 180}},
			},
			{
				NodeIdx:   12,
				ParamName: "vec_b",
				ValueEnum: ValueEnum{Vector: &VectorValue{
					X: 1.5,
					Z: 1.5,
				}},
			},
			{
				NodeIdx:   2,
				ParamName: "start_angle",
				ValueEnum: ValueEnum{Scalar: &ScalarValue{X: 180}},
			},
			{
				NodeIdx:   1,
				ParamName: "pos",
				ValueEnum: ValueEnum{Vector: &VectorValue{}},
			},
			{
				ParamName: "normal",
				ValueEnum: ValueEnum{Vector: &VectorValue{Z: 1}},
			},
		}},
	},
}
