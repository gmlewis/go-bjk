// Package ast defines the AST grammar for BJK files.
// See: https://github.com/setzer22/blackjack
package ast

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Lexer represents a lexer for the BJK grammar.
var Lexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Header", `(?:// BLACKJACK_VERSION_HEADER)[ ]*`},
	{"Ident", `[a-zA-Z]\w*`},
	{"Float", `\-?(?:\d*)?\.\d+`},
	{"Int", `\-?(?:\d*)?\d+`},
	{"String", `\"[^\"]*\"`},
	{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	{"Whitespace", `[ \t\n\r]+`},
})

// Parser represents a participle parser for the BJK grammar.
var Parser = participle.MustBuild[BJK](
	participle.Lexer(Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote("String"),
)

// BJK is the root of a Blackjack file.
type BJK struct {
	Version Version `Header @@`

	Graph *Graph `"(" @@* ")"`
}

// Graph represents the content of the Blackjack file.
type Graph struct {
	Nodes []*Node `"nodes" ":" "[" ( "(" @@* ")" ","? )* "]" ","?`

	DefaultNode *uint64 `( "default_node" ":" ( "Some" "(" @Int ")" | "None" ) ","? )?`

	UIData *UIData `( "ui_data" ":" ( "Some" "(" @@ ")" | "None" ) ","? )?`

	ExternalParameters *ExternalParameters `( "external_parameters" ":"  ( "Some" "(" @@ ")" | "None" ) ","? )?`
}

// Version represents the version of the Blackjack file.
type Version struct {
	Major int `@Int`
	Minor int `@Int`
	Patch int `@Int`
}

// Node represents a node in Blackjack.
type Node struct {
	OpName      string    `"op_name" ":" @String ","?`
	ReturnValue *string   `"return_value" ":" ( "Some" "(" @String ")" | "None" ) ","?`
	Inputs      []*Input  `"inputs" ":" "[" ( "(" @@* ")" ","? )* "]" ","?`
	Outputs     []*Output `"outputs" ":" "[" ( "(" @@* ")" ","? )* "]" ","?`
}

// Input represents a node's input.
type Input struct {
	Name     string         `"name" ":" @String ","*`
	DataType string         `"data_type" ":" @String ","*`
	Kind     DependencyKind `"kind" ":" @@ ","*`

	// Props are not preserved in the BJK file.
	Props map[string]any
}

// DependencyKind is an enum that represents an input's dependency.
// It is either an External or a Connection, but not both.
type DependencyKind struct {
	External   *External   `  "External" "(" @@* ")" ","?`
	Connection *Connection `| "Conection" "(" @@* ")" ","?`
}

// External represents an external dependency kind.
type External struct {
	Promoted *string `"promoted" ":" ( "Some" "(" @String ")" | "None" ) ","?`
}

// Connection represents a DependencyKind's connection.
type Connection struct {
	NodeIdx   uint64 `"node_idx" ":" @Int ","?`
	ParamName string `"param_name" ":" @String ","?`
}

// Output represents a node's output.
type Output struct {
	Name     string `"name" ":" @String ","*`
	DataType string `"data_type" ":" @String ","*`
}

// UIData represents data to drive the user interface.
type UIData struct {
	NodePositions    []*Vec2  `"(" "node_positions" ":" "[" @@* "]" ","?`
	NodeOrder        []uint64 `"node_order" ":" "[" ( @Int ","? )* "]" ","?`
	Pan              Vec2     `"pan" ":" @@ ","?`
	Zoom             float64  `"zoom" ":" @Float ","?`
	LockedGizmoNodes []uint64 `"locked_gizmo_nodes" ":" "["  ( @Int ","? )* "]" ","? ")" ","?`
}

// Vec2 represents a 2D vector (or point).
type Vec2 struct {
	X float64 `"(" @Float ","`
	Y float64 `@Float ")" ","?`
}

// ExternalParameters represents external parameters.
type ExternalParameters struct {
	ParamValues []*ParamValue `"(" "param_values" ":" "{" @@* "}" ","? ")" ","?`
}

// ParamValue is an enum that represents a parameter value.
// It is exactly one of the values.
type ParamValue struct {
	NodeIdx   uint64 `"(" "node_idx" ":" @Int ","?`
	ParamName string `"param_name" ":" @String ","? ")" ":"`

	ValueEnum ValueEnum `@@`
}

// ValueEnum represents an enum for a ParamValue.
type ValueEnum struct {
	Scalar    *ScalarValue    `  @@`
	Selection *SelectionValue `| @@`
	String    *StringValue    `| @@`
	Vector    *VectorValue    `| @@`
}

// VectorValue is one type of ParamValue.
type VectorValue struct {
	X float64 `"Vector" "(" "(" @Float ","`
	Y float64 `@Float ","`
	Z float64 `@Float ")" ")" ","?`
}

// ScalarValue is one type of ParamValue.
type ScalarValue struct {
	X float64 `"Scalar" "(" @Float ")" ","?`
}

// StringValue is one type of ParamValue.
type StringValue struct {
	S string `"String" "(" @String ")" ","?`
}

// SelectionValue is one type of ParamValue.
type SelectionValue struct {
	Selection string `"String" "(" @String ")" ","?`
}
