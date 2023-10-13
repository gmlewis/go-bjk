package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

// Node represents a Blackjack Node.
type Node struct {
}

func (c *Client) List() (map[string]*Node, error) {
	result := map[string]*Node{}

	if err := c.ls.DoString(`local N = require("node_library")
local nodes = {}
for k, v in pairs(N:listNodes()) do
    table.insert(nodes, v)
end
return nodes
`); err != nil {
		return nil, err
	}

	lv := c.ls.Get(-1) // get the value at the top of the stack
	if tbl, ok := lv.(*lua.LTable); ok {
		// fmt.Println(c.ls.ObjLen(tbl))
		tbl.ForEach(func(_, v lua.LValue) {
			result[v.String()] = &Node{}
		})
	}

	return result, nil
}
