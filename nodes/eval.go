// -*- compile-command: "go run ../cmd/make-scalar/main.go"; -*-

package nodes

import (
	"errors"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
)

// Eval "evaluates" a BJK design using lua.
func (c *Client) Eval(design *ast.BJK) error {
	if design == nil || design.Graph == nil || len(design.Graph.Nodes) == 0 {
		return errors.New("design missing nodes")
	}
	nodes := design.Graph.Nodes

	// Generate a lookup table for external parameters. key="nodeIdx,paramName" (e.g. "0,center")
	c.extParamsLookup = map[string]*ast.ValueEnum{}
	for _, pv := range design.Graph.ExternalParameters.ParamValues {
		key := genKey(int(pv.NodeIdx), pv.ParamName)
		c.extParamsLookup[key] = &pv.ValueEnum
	}

	// assume that the very last node is the "active" node.
	targetNodeIdx := len(nodes) - 1
	if err := c.runNode(nodes, targetNodeIdx); err != nil {
		return err
	}

	/*
			expr := fmt.Sprintf(`local N = require("node_library")
		local node = N:getNode(%q)
		print(node.inputs)
		print(node)
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
	*/
	return nil
}

func genKey(nodeIdx int, paramName string) string {
	return fmt.Sprintf("%v,%v", nodeIdx, paramName)
}

func (c *Client) runNode(nodes []*ast.Node, targetNodeIdx int) error {
	if targetNodeIdx >= len(nodes) {
		return fmt.Errorf("Eval: bad target node index %v, want 0..%v", targetNodeIdx, len(nodes))
	}
	log.Printf("runNode(%v)", targetNodeIdx)

	targetNode := nodes[targetNodeIdx]
	inputsTable := c.ls.NewTable()

	for _, input := range targetNode.Inputs {
		if input.Kind.External != nil {
			ve, ok := c.extParamsLookup[genKey(targetNodeIdx, input.Name)]
			if !ok {
				return fmt.Errorf("runNode(targetNodeIdx=%v), cannot find exteral param %q", targetNodeIdx, input.Name)
			}
			log.Printf("runNode: external ValueEnum=%#v", *ve)
			continue
		}
		if conn := input.Kind.Connection; conn != nil {
			log.Printf("runNode: connection from (%v,%v) to input node %v", conn.NodeIdx, conn.ParamName, input.Name)
			return fmt.Errorf("TODO: Call runNode recursively to calculate input and add to inputsTable")
			continue
		}
		// At this point, this input node has neither an extern parameter setting nor a connection - get the default value.
		log.Printf("runNode: TODO: GET DEFAULT VALUE FOR (%v,%v)", targetNodeIdx, input.Name)
		log.Printf("c.Nodes[%v]=%#v", targetNode.OpName, targetNode)
		log.Printf("input=%#v", input)
		log.Printf("input.Props=%#v", input.Props)
		defVal, ok := input.Props["default"]
		if !ok {
			return fmt.Errorf("runNode: input.Props['default'] could not be found: %#v", input.Props)
		}
		lVal, ok := defVal.(lua.LValue)
		if !ok {
			return fmt.Errorf("runNode: input.Props['default'] is type %T, want LValue", defVal)
		}
		inputsTable.RawSet(lua.LString(input.Name), lVal)
	}

	log.Printf("runNode: ALL INPUTS ARE RESOLVED - executing function %v.op(inputs)", targetNode.OpName)

	// top := c.ls.GetTop()
	expr := fmt.Sprintf("return require('node_library'):getNode('%v').op", targetNode.OpName)
	if err := c.ls.DoString(expr); err != nil {
		return err
	}
	// newTop := c.ls.GetTop()
	// log.Printf("before push: top=%v, newTop=%v", top, newTop)
	c.ls.Push(inputsTable)
	// newTop2 := c.ls.GetTop()
	// log.Printf("after push: newTop2=%v", newTop2)
	c.ls.Call(1, 1)
	outputs := c.ls.CheckTable(1)
	if outputs == nil {
		return fmt.Errorf("runNode: expected outputs table, got type %v: %v", c.ls.Get(1).Type(), c.ls.Get(1).String())
	}

	outputs.ForEach(func(k, v lua.LValue) {
		log.Printf("outputs[%q] = %v", k, v)
	})

	return nil
}
