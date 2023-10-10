// Package ast defines the AST grammar for BJK files.
package ast

// BJK is the root of a Blackjack file.
// See: https://github.com/setzer22/blackjack
type BJK struct {
	Version Version `"// BLACKJACK_VERSION_HEADER" @@`
	Graph   Graph   `"(" @@ ")"`
}

// Graph represents the content of the Blackjack file.
type Graph struct {
	Nodes []*Node `"nodes:" "[" @@ "]" ","*`

	DefaultNode *uint64 `"default_node:" "Some(" @Int ")" ","*`

	UIData *UIData

	ExternalParameters *ExternalParameters
}

// Version represents the version of the Blackjack file.
type Version struct {
	Major int `@Int`
	Minor int `@Int`
	Patch int `@Int`
}

// Node represents a node in Blackjack.
type Node struct {
	OpName      string    `"op_name:" "\"" @Ident "\"" ","*`
	ReturnValue *string   `"return_value:" "Some(" @Ident ")" ","*`
	Inputs      []*Input  `"inputs:" "[" @@ "]" ","*`
	Outputs     []*Output `"outputs:" "[" @@ "]" ","*`
}

// Input represents a node's input.
type Input struct {
	Name     string         `"name:" @Ident ","*`
	DataType string         `"data_type:" @Ident ","*`
	Kind     DependencyKind `"kind:" @@ ","*`
}

// DependencyKind is an enum that represents an input's dependency.
// It is either an External or a Connection, but not both.
type DependencyKind struct {
	External   *External   `"External(" @@ ")" ","*`
	Connection *Connection `"Connection(" @@ ")" ","*`
}

// External represents an external dependency kind.
type External struct {
	Promoted *string `"Some(" @Ident ")"`
}

// Connection represents a DependencyKind's connection.
type Connection struct {
	NodeIdx   uint64 `"node_idx:" @Int ","*`
	ParamName string `"param_name:" @Ident ","*`
}

// Output represents a node's output.
type Output struct {
	Name     string `"name:" @Ident ","*`
	DataType string `"data_type:" @Ident ","*`
}

// UIData represents data to drive the user interface.
type UIData struct {
	NodePositions    []*Vec2
	NodeOrder        []uint64
	Pan              Vec2
	Zoom             float64
	LockedGizmoNodes []uint64
}

// Vec2 represents a 2D vector (or point).
type Vec2 struct {
	X float64
	Y float64
}

// ExternalParameters represents external parameters.
type ExternalParameters struct {
	ParamValues []*ParamValue
}

// ParamValue is an enum that represents a parameter value.
// It is exactly one of the values.
type ParamValue struct {
	NodeIdx   uint64
	ParamName string

	// Enum with exactly one of the following values:
	Vector    *VectorValue
	Scalar    *ScalarValue
	String    *StringValue
	Selection *SelectionValue
}

// VectorValue is one type of ParamValue.
type VectorValue struct {
	X float64
	Y float64
	Z float64
}

// ScalarValue is one type of ParamValue.
type ScalarValue struct {
	X float64
}

// StringValue is one type of ParamValue.
type StringValue struct {
	S string
}

// SelectionValue is one type of ParamValue.
type SelectionValue struct {
	Selection string
}
