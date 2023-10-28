// -*- compile-command: "go run ../cmd/make-box/main.go"; -*-

package nodes

import (
	"errors"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/ast"
)

// Eval "evaluates" a BJK design using lua.
func (c *Client) Eval(design *ast.BJK) error {
	if design == nil || design.Graph == nil {
		return errors.New("design missing graph")
	}

	// assume that the very last node is the "active" node.
	nodes := design.Graph.Nodes
	lastNode := nodes[len(nodes)-1]

	expr := fmt.Sprintf(`local N = require("node_library")
local node = N:getNode(%q)
return node.op(node.inputs)
`, lastNode.OpName)
	if err := c.ls.DoString(expr); err != nil {
		return err
	}
	defer c.ls.Pop(1)
	lv := c.ls.Get(-1) // get the value at the top of the stack
	log.Printf("lv=%#v", lv)

	log.Printf("Node(%q) has %v inputs", lastNode.OpName, len(lastNode.Inputs))
	c.debug = true
	c.showTop()
	return nil
}
