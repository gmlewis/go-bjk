package nodes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gmlewis/go-bjk/ast"
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
	}
	b.NodeOrder = append(b.NodeOrder, name)

	return b
}

// Connect connects the `from` node.output to the `to` node.input.
func (b *Builder) Connect(from, to string) *Builder {
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
