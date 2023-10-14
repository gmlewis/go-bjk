package nodes

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/hexops/valast"
)

// Builder represents a BJK builder.
type Builder struct {
	c    *Client
	errs []error

	Nodes     map[string]*ast.Node
	NodeOrder []string
}

// NewBuilder returns a new BJK Builder.
func (c *Client) NewBuilder() *Builder {
	return &Builder{c: c, Nodes: map[string]*ast.Node{}}
}

// AddNode adds a new node to the design with the optional args.
func (b *Builder) AddNode(name string, args ...string) *Builder {
	if name == "" {
		b.errs = append(b.errs, errors.New("AddNode: name cannot be empty"))
		return b
	}

	parts := strings.Split(name, ".")
	n, ok := b.c.Nodes[parts[0]]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("AddNode: unknown node type '%v'", parts[0]))
		return b
	}

	if _, ok := b.Nodes[name]; ok {
		b.errs = append(b.errs, fmt.Errorf("AddNode: node '%v' already exists", name))
		return b
	}

	b.Nodes[name] = &ast.Node{
		OpName:      parts[0],
		ReturnValue: n.ReturnValue,
		Inputs:      n.Inputs,
		Outputs:     n.Outputs,

		Label:     n.Label,
		NodeIndex: uint64(len(b.NodeOrder)), // 0-based indices
	}
	b.NodeOrder = append(b.NodeOrder, name)

	return b
}

// Connect connects the `from` node.output to the `to` node.input.
func (b *Builder) Connect(from, to string) *Builder {
	fromParts := strings.Split(from, ".")
	if len(fromParts) != 3 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'from' name: %[1]q want 3 parts, got %v", from, to, len(fromParts)))
		return b
	}

	fromNodeName := fmt.Sprintf("%v.%v", fromParts[0], fromParts[1])
	fromNode, ok := b.Nodes[fromNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node: %q", from, to, fromNodeName))
		return b
	}

	fromOutput, ok := fromNode.GetOutput(fromParts[2])
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'from' node output: %q", from, to, fromParts[2]))
		return b
	}

	toParts := strings.Split(to, ".")
	if len(toParts) != 3 {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q): unable to parse 'to' name: %[1]q want 3 parts, got %v", from, to, len(toParts)))
		return b
	}

	toNodeName := fmt.Sprintf("%v.%v", toParts[0], toParts[1])
	toNode, ok := b.Nodes[toNodeName]
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node: %q", from, to, toNodeName))
		return b
	}

	toInput, ok := toNode.GetInput(toParts[2])
	if !ok {
		b.errs = append(b.errs, fmt.Errorf("Connect(%q,%q) unable to find 'to' node input: %q", from, to, toParts[2]))
		return b
	}

	if b.c.debug {
		log.Printf("Connect(%q,%q):\nfrom:\n%#v\nto:\n%#v", from, to, valast.String(fromOutput), valast.String(toInput))
	}

	toInput.Kind.Connection = &ast.Connection{
		NodeIdx:   fromNode.NodeIndex,
		ParamName: fromOutput.Name,
	}

	return b
}

// Builder builds the design and returns the result.
func (b *Builder) Build() (*ast.BJK, error) {
	if len(b.errs) > 0 {
		return nil, errors.Join(b.errs...)
	}

	bjk := ast.New()
	if len(b.NodeOrder) == 0 {
		return bjk, nil
	}

	g := bjk.Graph

	for _, k := range b.NodeOrder {
		g.Nodes = append(g.Nodes, b.Nodes[k])
	}

	dn := uint64(len(b.NodeOrder) - 1)
	g.DefaultNode = &dn

	return bjk, nil
}
