package nodes

import (
	"fmt"
	"log"

	"github.com/gmlewis/go-bjk/ast"
	lua "github.com/yuin/gopher-lua"
)

func (c *Client) list() (map[string]*ast.Node, error) {
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
			if c.debug {
				log.Printf("list: k=%v,v=%v, k=%v,v=%#v", k.Type(), v.Type(), k, v)
			}
			node, err := c.luaToNode(k.String(), v)
			if err != nil {
				return
			}
			result[k.String()] = node
		})
	}

	return result, nil
}

func (c *Client) luaToNode(nodeName string, lv lua.LValue) (*ast.Node, error) {
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
		OpName:      nodeName,
		ReturnValue: returnValue,
		Inputs:      inputs,
		Outputs:     outputs,
		Label:       t.RawGetString("label").String(),
	}
	return node, nil
}

func (c *Client) luaToInput(lv lua.LValue) (*ast.Input, error) {
	t, ok := lv.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("luaToInput: expected LTable, got %v", lv.Type())
	}

	if c.debug {
		log.Printf("luaToInput: t=%#v", t)
	}

	props := map[string]lua.LValue{}
	t.ForEach(func(k, v lua.LValue) {
		if c.debug {
			log.Printf("luaToInput: props[%v]=%#v", k, v)
		}
		props[k.String()] = v
	})

	name := t.RawGetString("name").String()
	dataType := t.RawGetString("type").String()
	input := &ast.Input{
		Name:     name,
		DataType: dataType,
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
