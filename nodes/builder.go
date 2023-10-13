package nodes

import "github.com/gmlewis/go-bjk/ast"

// Builder represents a BJK builder.
type Builder struct {
	c *Client
}

// NewBuilder returns a new BJK Builder.
func (c *Client) NewBuilder() *Builder {
	return &Builder{c: c}
}

// AddNode adds a new node to the design with the optional args.
func (b *Builder) AddNode(name string, args ...string) *Builder {
	return b
}

// Connect connects the `from` node.output to the `to` node.input.
func (b *Builder) Connect(from, to string) *Builder {
	return b
}

// Builder builds the design and returns the result.
func (b *Builder) Build() (*ast.BJK, error) {
	result := ast.New()
	return result, nil
}
