// -*- compile-command: "go run ../cmd/make-box/main.go"; -*-

package nodes

import (
	"errors"
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/exp/maps"
)

// Eval "evaluates" a BJK design using lua and returns a Mesh if one was generated.
func (c *Client) Eval(design *ast.BJK) (*Mesh, error) {
	if design == nil || design.Graph == nil || len(design.Graph.Nodes) == 0 {
		return nil, errors.New("design missing nodes")
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
		return nil, err
	}

	targetNode := nodes[targetNodeIdx]
	outMesh, ok := targetNode.EvalOutputs["out_mesh"]
	if !ok {
		log.Printf("WARNING: node %q missing output 'out_mesh', choices are: %+v", targetNode.OpName, maps.Keys(targetNode.EvalOutputs))
		return nil, nil
	}
	ud, ok := outMesh.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("'out_mesh' of type %T, expected *LUserData", outMesh)
	}
	mesh, ok := ud.Value.(*Mesh)
	if !ok {
		return nil, fmt.Errorf("'out_mesh' LUserData of type %T, expected *Mesh", ud.Value)
	}

	return mesh, nil
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
	if targetNode.EvalOutputs != nil {
		return nil // this node has already been evaluated.
	}
	targetNode.EvalOutputs = map[string]lua.LValue{}

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
			if err := c.runNode(nodes, int(conn.NodeIdx)); err != nil {
				return err
			}
			lVal, ok := nodes[conn.NodeIdx].EvalOutputs[conn.ParamName]
			if !ok {
				return fmt.Errorf("runNode(targetNodeIdx=%v), cannot find node[%v]('%v') output param %q, choices are: %+v", targetNodeIdx, conn.NodeIdx, nodes[conn.NodeIdx].OpName, conn.ParamName, maps.Keys(nodes[conn.NodeIdx].EvalOutputs))
			}
			inputsTable.RawSet(lua.LString(input.Name), lVal)
			continue
		}
		// At this point, this input node has neither an extern parameter setting nor a connection - get the default value.
		log.Printf("runNode: TODO: GET DEFAULT VALUE FOR (%v,%v)", targetNodeIdx, input.Name)
		log.Printf("c.Nodes[%v]=%#v", targetNode.OpName, targetNode)
		log.Printf("input=%#v", input)
		log.Printf("input.Props=%#v", input.Props)

		var lVal lua.LValue
		var ok bool
		switch input.DataType {
		case "enum":
			selected, ok := input.Props["selected"].(lua.LNumber)
			if !ok {
				return fmt.Errorf("runNode: input.Props['selected'] enum expected lua.LNumber, got %T: %#v", input.Props["selected"], input)
			}
			values, ok := input.Props["values"].(*lua.LTable)
			if !ok {
				return fmt.Errorf("runNode: input.Props['values'] enum expected *lua.LTable, got %T: %#v", input.Props["values"], input)
			}
			lVal = values.RawGet(selected + 1) // Lua is 1-indexed, but Blackjack is 0-indexed!
			if lVal.String() == "nil" {
				log.Fatalf("programming error: values.RawGet=(%v,%v), values=%#v", lVal.String(), lVal.Type(), values)
			}
			log.Printf("values.RawGet=(%v,%v), values=%#v", lVal.String(), lVal.Type(), values)
		default:
			if lVal, ok = input.Props["default"]; !ok {
				return fmt.Errorf("runNode: input.Props['default'] could not be found: %#v", input.Props)
			}
		}

		if lVal == nil {
			return fmt.Errorf("programming error: lVal remains unset for input %#v", *input)
		}
		inputsTable.RawSet(lua.LString(input.Name), lVal)
	}

	log.Printf("runNode: ALL INPUTS ARE RESOLVED - executing function %v.op(inputs)", targetNode.OpName)

	top := c.ls.GetTop()
	expr := fmt.Sprintf("return require('node_library'):getNode('%v').op", targetNode.OpName)
	if err := c.ls.DoString(expr); err != nil {
		return err
	}
	newTop := c.ls.GetTop()
	log.Printf("before push: top=%v, newTop=%v", top, newTop)
	c.ls.Push(inputsTable)
	newTop2 := c.ls.GetTop()
	log.Printf("after push: newTop2=%v", newTop2)
	c.ls.Call(1, 1)
	newTop3 := c.ls.GetTop()
	log.Printf("after push: newTop3=%v", newTop3)
	outputs := c.ls.CheckTable(1)
	if outputs == nil {
		return fmt.Errorf("runNode: expected outputs table, got type %v: %v", c.ls.Get(1).Type(), c.ls.Get(1).String())
	}
	log.Printf("lua execution returned table: %#v", *outputs)

	outputs.ForEach(func(k, v lua.LValue) {
		targetNode.EvalOutputs[k.String()] = v
		log.Printf("outputs[%q] = %v", k, v)
	})
	c.ls.Pop(1) // remove returned table from lua stack

	// Now verify that all the expected outputs have been assigned:
	for _, output := range targetNode.Outputs {
		if _, ok := targetNode.EvalOutputs[output.Name]; !ok {
			log.Fatalf("execution of node '%v' failed to generate expected output name '%v'! Aborting.", targetNode.OpName, output.Name)
		}
	}

	return nil
}
