// -*- compile-command: "go run ../cmd/make-box/main.go"; -*-

package nodes

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

	targetNodeIdx := len(nodes) - 1
	if design.Graph.DefaultNode != nil {
		targetNodeIdx = int(*design.Graph.DefaultNode)
	}
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
	if c.debug {
		log.Printf("runNode(%v)", targetNodeIdx)
	}

	targetNode := nodes[targetNodeIdx]
	if targetNode.EvalOutputs != nil {
		return nil // this node has already been evaluated.
	}
	targetNode.EvalOutputs = map[string]lua.LValue{}

	inputsTable := c.ls.NewTable()

	nameToKey, err := c.genNumToKeyMap(targetNode)
	if err != nil {
		return fmt.Errorf("genNumToKeyMap: %w", err)
	}

	for _, input := range targetNode.Inputs {
		if input.Props == nil {
			input.Props = map[string]lua.LValue{}
		}

		if input.Kind.External != nil {
			ve, ok := c.extParamsLookup[genKey(targetNodeIdx, input.Name)]
			if !ok {
				if input.DataType != "BJK_MESH" {
					return fmt.Errorf("runNode(targetNodeIdx=%v), cannot find external param %q", targetNodeIdx, input.Name)
				}
				input.Props[input.Name] = lua.LNil
				inputsTable.RawSet(lua.LString(input.Name), lua.LNil)
				if c.debug {
					log.Printf("Setting node %q input %q to nil", targetNode.OpName, input.Name)
				}
				continue
			}
			if c.debug {
				log.Printf("runNode: external ValueEnum=%#v", *ve)
			}
			lval := valueEnumToLValue(c.ls, ve)
			input.Props[input.Name] = lval
			inputsTable.RawSet(lua.LString(input.Name), lval)
			if c.debug {
				log.Printf("Setting node %q input %q to %v", targetNode.OpName, input.Name, lval)
			}
			// TODO: honor properties like min, max, soft_max, default, num_decimals, etc?
			if _, ok := nameToKey[input.Name]; !ok {
				log.Printf("WARNING! setting lua input %q on node %q but it is no longer declared as one of its inputs!", input.Name, targetNode.OpName)
			}
			continue
		}
		if conn := input.Kind.Connection; conn != nil {
			if c.debug {
				log.Printf("runNode: connection from (%v,%v) to input node %v", conn.NodeIdx, conn.ParamName, input.Name)
			}
			if err := c.runNode(nodes, int(conn.NodeIdx)); err != nil {
				return err
			}
			lVal, ok := nodes[conn.NodeIdx].EvalOutputs[conn.ParamName]
			if !ok {
				return fmt.Errorf("runNode(targetNodeIdx=%v), cannot find node[%v]('%v') output param %q, choices are: %+v", targetNodeIdx, conn.NodeIdx, nodes[conn.NodeIdx].OpName, conn.ParamName, maps.Keys(nodes[conn.NodeIdx].EvalOutputs))
			}
			inputsTable.RawSet(lua.LString(input.Name), lVal)
			if c.debug {
				log.Printf("Setting node %q input %q to %v", targetNode.OpName, input.Name, lVal)
			}
			// TODO: honor properties like min, max, soft_max, default, num_decimals, etc?
			if _, ok := nameToKey[input.Name]; !ok {
				log.Printf("WARNING! setting lua input %q on node %q but it is no longer declared as one of its inputs!", input.Name, targetNode.OpName)
			}
			continue
		}
		// At this point, this input node has neither an extern parameter setting nor a connection - get the default value.
		if c.debug {
			log.Printf("c.Nodes[%v]=%#v", targetNode.OpName, targetNode)
			log.Printf("input=%#v", input)
			log.Printf("input.Props=%#v", input.Props)
		}

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
			if c.debug {
				log.Printf("values.RawGet=(%v,%v), values=%#v", lVal.String(), lVal.Type(), values)
			}
		case "mesh":
			lVal = lua.LNil
		default:
			if lVal, ok = input.Props["default"]; !ok {
				return fmt.Errorf("runNode: input.Props['default'] could not be found: %#v", input.Props)
			}
		}

		if lVal == nil {
			return fmt.Errorf("programming error: lVal remains unset for input %#v", *input)
		}
		inputsTable.RawSet(lua.LString(input.Name), lVal)
		if c.debug {
			log.Printf("Setting node %q input %q to %v", targetNode.OpName, input.Name, lVal)
		}
		// TODO: honor properties like min, max, soft_max, default, num_decimals, etc?
		if _, ok := nameToKey[input.Name]; !ok {
			log.Printf("WARNING! setting lua input %q on node %q but it is no longer declared as one of its inputs!", input.Name, targetNode.OpName)
		}
	}

	if c.debug {
		log.Printf("runNode: ALL INPUTS ARE RESOLVED - executing function %v.op(inputs)", targetNode.OpName)
	}

	expr := fmt.Sprintf("return require('node_library'):getNode('%v').op", targetNode.OpName)
	if err := c.ls.DoString(expr); err != nil {
		return err
	}
	c.ls.Push(inputsTable)
	c.ls.Call(1, 1)
	outputs := c.ls.CheckTable(1)
	if outputs == nil {
		return fmt.Errorf("runNode: expected outputs table, got type %v: %v", c.ls.Get(1).Type(), c.ls.Get(1).String())
	}
	if c.debug {
		log.Printf("lua execution returned table: %#v", *outputs)
	}

	outputs.ForEach(func(k, v lua.LValue) {
		targetNode.EvalOutputs[k.String()] = v
		if c.debug {
			log.Printf("outputs[%q] = %v", k, v)
		}
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

// valueEnumToLValue converts an ast.ValueEnum to a lua.LValue.
func valueEnumToLValue(ls *lua.LState, ve *ast.ValueEnum) lua.LValue {
	switch {
	case ve.Scalar != nil:
		return lua.LNumber(ve.Scalar.X)
	case ve.Selection != nil:
		return lua.LString(ve.Selection.Selection)
	case ve.StrVal != nil:
		return lua.LString(ve.StrVal.S)
	case ve.Vector != nil:
		vec3 := &Vec3{X: ve.Vector.X, Y: ve.Vector.Y, Z: ve.Vector.Z}
		return newVec3LValue(ls, vec3)
	default:
		log.Fatalf("programming error: eval.go: valueEnumToLValue: unhandled ValueEnum: %#v", *ve)
	}
	return lua.LNumber(0)
}

func (c *Client) genNumToKeyMap(targetNode *ast.Node) (map[string]string, error) {
	expr := fmt.Sprintf("return require('node_library'):getNode('%v').inputs", targetNode.OpName)
	if err := c.ls.DoString(expr); err != nil {
		return nil, err
	}
	defer c.ls.Pop(1) // remove returned table from lua stack

	inputsTable := c.ls.CheckTable(1)
	if inputsTable == nil {
		return nil, fmt.Errorf("runNode: expected outputs table, got type %v: %v", c.ls.Get(1).Type(), c.ls.Get(1).String())
	}
	if c.debug {
		log.Printf("lua execution returned inputs table: %#v", *inputsTable)
	}

	nameToKey := make(map[string]string, inputsTable.Len())
	keyToDefaultLVals := make(map[string]lua.LValue, inputsTable.Len())
	inputsTable.ForEach(func(k, v lua.LValue) {
		if t, ok := v.(*lua.LTable); ok {
			t.ForEach(func(k2, v2 lua.LValue) {
				if k2.String() == "name" {
					nameToKey[v2.String()] = k.String()
				} else if k2.String() == "default" {
					keyToDefaultLVals[k.String()] = v2
				}
				if c.debug {
					log.Printf("inputsTable[%q][%q] = %T: %v", k, k2, v2, v2)
				}
			})
		}
	})
	if c.debug {
		log.Printf("nameToKey map: %+v", nameToKey)
	}
	// set default values in case this BJK is out-of-date with the actual BJK Node.
	for name, key := range nameToKey {
		if defLVal, ok := keyToDefaultLVals[key]; ok {
			if c.debug {
				log.Printf("Setting node %q input %q to default value %v", targetNode.OpName, name, defLVal)
			}
			inputsTable.RawSet(lua.LString(name), defLVal)
		}
	}

	return nameToKey, nil
}

// GetScalar gets the value of a scalar from a design and returns it.
func (c *Client) GetScalar(design *ast.BJK, nodeName string) (float64, error) {
	if design == nil || design.Graph == nil {
		return 0, errors.New("design missing graph")
	}

	if c.cachedMesh == nil {
		mesh, err := c.Eval(design)
		if err != nil {
			return 0, err
		}
		c.cachedMesh = mesh
	}

	parts := strings.Split(nodeName, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("GetScalar - want nodeName in 2 parts, got: %+v", parts)
	}
	opName, outputName := parts[0], parts[1]
	for _, n := range design.Graph.Nodes {
		if n.OpName == opName {
			v, ok := n.EvalOutputs[outputName]
			if ok {
				return float64(lua.LVAsNumber(v)), nil
			}
		}
	}

	return 0, fmt.Errorf("output node '%v' not found", nodeName)
}
