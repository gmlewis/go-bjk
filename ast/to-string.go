package ast

import (
	"fmt"
	"strings"
)

const indent = "    "

func indentBlock(s string, f func(fmtStr string, args ...any)) {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i < len(lines)-1 {
			f(indent + line)
			continue
		}
		f(indent + line + ",")
	}
}

func (b *BJK) String() string {
	v := b.Version
	return fmt.Sprintf("%v %v %v %v\n%v", headerStr, v.Major, v.Minor, v.Patch, b.Graph)
}

func (g *Graph) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }
	f("nodes: [")

	for _, n := range g.Nodes {
		indentBlock(n.String(), f)
	}

	f("],")

	if g.DefaultNode == nil {
		f("default_node: None,")
	} else {
		f("default_node: Some(%v),", *g.DefaultNode)
	}

	if g.UIData == nil {
		f("ui_data: None,")
	} else {
		f("ui_data: Some((")
		indentBlock(g.UIData.String(), f)
		f(")),")
	}

	if g.ExternalParameters == nil {
		f("external_parameters: None,")
	} else {
		f("external_parameters: Some((")
		indentBlock(g.ExternalParameters.String(), f)
		f(")),")
	}

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (n *Node) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }

	f("op_name: %q,", n.OpName)

	if n.ReturnValue != nil {
		f("return_value: Some(%q),", *n.ReturnValue)
	} else {
		f("return_value: None,")
	}

	f("inputs: [")
	for _, in := range n.Inputs {
		indentBlock(in.String(), f)
	}
	f("],")

	f("outputs: [")
	for _, out := range n.Outputs {
		indentBlock(out.String(), f)
	}
	f("],")

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (in *Input) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }

	f("name: %q,", in.Name)
	f("data_type: %q,", dataTypeToBJK(in.DataType))

	if v := in.Kind.Connection; v != nil {
		f("kind: Conection(") // [sic]
		f(indent+"node_idx: %v,", v.NodeIdx)
		f(indent+"param_name: %q,", v.ParamName)
		f("),")
	} else if in.Kind.External != nil {
		f("kind: External(")
		// TODO - f(indent + "promoted: None,")
		f("),")
	} else {
		f("kind: External(")
		f(indent + "promoted: None,")
		f("),")
	}

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (out *Output) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }

	f("name: %q,", out.Name)
	f("data_type: %q,", dataTypeToBJK(out.DataType))

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func dataTypeToBJK(dt string) string {
	if v, ok := dataTypeLookup[dt]; ok {
		return v
	}
	return dt
}

var dataTypeLookup = map[string]string{
	"enum":   "BJK_STRING",
	"mesh":   "BJK_MESH",
	"scalar": "BJK_SCALAR",
	"vec3":   "BJK_VECTOR",
}

func (ui *UIData) String() string {
	var lines []string
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(fmtStr, args...)) }

	f("node_positions: [")
	for _, v2 := range ui.NodePositions {
		f(indent+"(%0.5f, %0.5f),", v2.X, v2.Y)
	}
	f("],")

	f("node_order: [")
	for _, idx := range ui.NodeOrder {
		f(indent+"%v,", idx)
	}
	f("],")

	f("pan: (%0.5f, %0.5f),", ui.Pan.X, ui.Pan.Y)
	f("zoom: %0.7f,", ui.Zoom)
	f("locked_gizmo_nodes: []") // trailing comma added by indentBlock
	return strings.Join(lines, "\n")
}

func (ep *ExternalParameters) String() string {
	var lines []string
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(fmtStr, args...)) }

	f("param_values: {")
	for _, pv := range ep.ParamValues {
		indentBlock(pv.String(), f)
	}
	f("}") // trailing comma added by indentBlock

	return strings.Join(lines, "\n")
}

func (pv *ParamValue) String() string {
	var lines []string
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(fmtStr, args...)) }

	f("(")
	f(indent+"node_idx: %v,", pv.NodeIdx)
	f(indent+"param_name: %q,", pv.ParamName)
	f("): %v", pv.ValueEnum) // trailing comma added by indentBlock

	return strings.Join(lines, "\n")
}

func (ev ValueEnum) String() string {
	// Only one value should be non-nil, so just combine them all:
	return fmt.Sprintf("%v%v%v%v",
		ev.Scalar,
		ev.Selection,
		ev.StrVal,
		ev.Vector)
}

func (sv *ScalarValue) String() string {
	if sv == nil {
		return ""
	}
	return fmt.Sprintf("Scalar(%v)", addDot0(sv.X))
}

func (sv *StringValue) String() string {
	if sv == nil {
		return ""
	}
	return fmt.Sprintf("String(%q)", sv.S)
}

func (sv *SelectionValue) String() string {
	if sv == nil {
		return ""
	}
	return fmt.Sprintf("Selection(%q)", sv.Selection)
}

func (vv *VectorValue) String() string {
	if vv == nil {
		return ""
	}
	return fmt.Sprintf("Vector((%v, %v, %v))", addDot0(vv.X), addDot0(vv.Y), addDot0(vv.Z))
}

func addDot0(f float64) string {
	s := fmt.Sprintf("%v", f)
	if strings.Contains(s, ".") {
		return s
	}
	return s + ".0"
}
