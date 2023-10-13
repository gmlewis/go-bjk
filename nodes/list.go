package nodes

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

// Node represents a Blackjack Node.
type Node struct {
}

func (c *Client) List() (map[string]*Node, error) {
	result := map[string]*Node{}

	if err := c.ls.DoString(`local N = require("node_library")
local nodes = {}
for _, v in pairs(N:listNodes()) do
    nodes[v] = N:getNode(v)
end
return nodes
`); err != nil {
		return nil, err
	}

	lv := c.ls.Get(-1) // get the value at the top of the stack
	if tbl, ok := lv.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			log.Printf("Node['%v'] = %#v", k, v)
			result[k.String()] = &Node{}
		})
	}

	return result, nil
}

/*
2023/10/13 10:06:48 Node['Helix'] = &lua.LTable{Metatable:(*lua.LNilType)(0x102877b80), array:[]lua.LValue(nil), dict:map[lua.LValue]lua.LValue(nil), strdict:map[string]lua.LValue{"inputs":(*lua.LTable)(0x14000488c60), "label":"Helix", "op":(*lua.LFunction)(0x1400048abc0), "outputs":(*lua.LTable)(0x14000489080), "returns":"out_mesh"}, keys:[]lua.LValue{"label", "op", "inputs", "outputs", "returns"}, k2i:map[lua.LValue]int{"inputs":2, "label":0, "op":1, "outputs":3, "returns":4}}
*/
