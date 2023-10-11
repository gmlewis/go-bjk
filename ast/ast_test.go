package ast

import (
	_ "embed"
	"os"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/google/go-cmp/cmp"
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
			want:  &BJK{Graph: &Graph{DefaultNode: Uint(15)}},
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
					DefaultNode: Uint(15),
					UIData: &UIData{
						NodePositions: []*Vec2{{X: 1072.3687, Y: 232.1065}, {X: 1070.6208, Y: 990.6514}},
						NodeOrder:     []uint64{2, 3},
						Pan:           Vec2{X: -0.6252365, Y: -565.1915},
						Zoom:          1.342995,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simpleLexer := lexer.MustSimple([]lexer.SimpleRule{
				{"Header", `(?:// BLACKJACK_VERSION_HEADER)[ ]*`},
				{"Ident", `[a-zA-Z]\w*`},
				{"Float", `\-?(?:\d*)?\.\d+`},
				{"Int", `\-?(?:\d*)?\d+`},
				{"String", `\"[^\"]*\"`},
				{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
				{"Whitespace", `[ \t\n\r]+`},
			})
			parser := participle.MustBuild[BJK](
				participle.Lexer(simpleLexer),
				participle.Elide("Whitespace"),
				participle.Unquote("String"),
				// participle.UseLookahead(20),
			)

			got, err := parser.ParseString("", tt.input, participle.Trace(os.Stderr))
			if err != nil {
				t.Logf("%v\n", tt.input)
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Logf("%v\n", tt.input)
				t.Errorf("ParseString mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func String(s string) *string { return &s }
func Uint(v uint64) *uint64   { return &v }
