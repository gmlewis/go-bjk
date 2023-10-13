package nodes

import (
	"fmt"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
)

func (c *Client) List() (map[string]*ast.Node, error) {
	result := map[string]*ast.Node{}

	if err := c.ls.DoString(`local N = require("node_library")
local nodes = {}
for _, v in pairs(N:listNodes()) do
    nodes[v] = N:getNode(v)
end
return nodes
`); err != nil {
		return nil, err
	}
	defer c.ls.Pop(1)

	lv := c.ls.Get(-1) // get the value at the top of the stack
	if tbl, ok := lv.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			node, err := c.luaToNode(v)
			if err != nil {
				return
			}
			result[k.String()] = node
		})
	}

	return result, nil
}

func (c *Client) luaToNode(lv lua.LValue) (*ast.Node, error) {
	t, ok := lv.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("luaToNode: expected LTable, got %v", lv.Type())
	}

	var returnValue *string
	if lv := t.RawGetString("returns"); lv.Type() == lua.LTString {
		v := lv.String()
		returnValue = &v
	}

	var inputs []*ast.Input
	if it, ok := t.RawGetString("inputs").(*lua.LTable); ok {
		var err error
		it.ForEach(func(_, v lua.LValue) {
			input, err2 := c.luaToInput(v)
			if err2 != nil {
				err = err2
			}
			inputs = append(inputs, input)
		})
		if err != nil {
			return nil, err
		}
	}

	var outputs []*ast.Output
	if ot, ok := t.RawGetString("outputs").(*lua.LTable); ok {
		var err error
		ot.ForEach(func(k, v lua.LValue) {
			output, err2 := c.luaToOutput(v)
			if err2 != nil {
				err = err2
			}
			outputs = append(outputs, output)
		})
		if err != nil {
			return nil, err
		}
	}

	node := &ast.Node{
		OpName:      t.RawGetString("label").String(),
		ReturnValue: returnValue,
		Inputs:      inputs,
		Outputs:     outputs,
	}
	return node, nil
}

func (c *Client) luaToInput(lv lua.LValue) (*ast.Input, error) {
	t, ok := lv.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("luaToInput: expected LTable, got %v", lv.Type())
	}

	props := map[string]any{}
	t.ForEach(func(k, v lua.LValue) {
		props[k.String()] = v
	})

	input := &ast.Input{
		Name:     t.RawGetString("name").String(),
		DataType: t.RawGetString("type").String(),
		Props:    props,
	}
	return input, nil
}

func (c *Client) luaToOutput(lv lua.LValue) (*ast.Output, error) {
	t, ok := lv.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("luaToOutput: expected LTable, got %v", lv.Type())
	}

	output := &ast.Output{
		Name:     t.RawGetString("name").String(),
		DataType: t.RawGetString("type").String(),
	}
	return output, nil
}
