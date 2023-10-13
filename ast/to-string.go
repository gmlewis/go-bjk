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
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (n *Node) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }

	f("op_name: %q,", n.OpName)

	if n.ReturnValue != nil {
		f("return_value: %q,", *n.ReturnValue)
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
	f("data_type: %q,", in.DataType)

	if in.Kind.External != nil {
		f("kind: External(")
		f(indent + "promoted: None")
		f("),")
	}

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func (out *Output) String() string {
	lines := []string{"("}
	f := func(fmtStr string, args ...any) { lines = append(lines, fmt.Sprintf(indent+fmtStr, args...)) }

	f("name: %q,", out.Name)
	f("data_type: %q,", out.DataType)

	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}
