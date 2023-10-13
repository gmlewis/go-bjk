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

	Nodes map[string]*ast.Node
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

	b.Nodes[name] = n

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

	result := ast.New()
	return result, nil
}
