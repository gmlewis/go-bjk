// -*- compile-command: "go run ../examples/bifilar-electromagnet/main.go -o '' -stl ../out.stl"; -*-

package nodes

import (
	"errors"

	"github.com/gmlewis/go-bjk/ast"
)

// ToSTL "renders" a BJK design to a binary STL byte slice.
func (c *Client) ToSTL(design *ast.BJK) ([]byte, error) {
	if design == nil || design.Graph == nil {
		return nil, errors.New("design missing graph")
	}

	if err := c.Eval(design); err != nil {
		return nil, err
	}

	// // assume that the very last node is the "active" node.
	// nodes := design.Graph.Nodes
	// lastNode := nodes[len(nodes)-1]
	//
	// log.Printf("Node(%q) has %v inputs", lastNode.OpName, len(lastNode.Inputs))
	// c.debug = true
	// c.showTop()
	return nil, nil
}
