package nodes

import (
	"fmt"
	"log"

	lua "github.com/yuin/gopher-lua"
)

// Node represents a Blackjack Node.
type Node struct {
}

func (c *Client) List() map[string]*Node {
	result := map[string]*Node{}

	// fn, ok := c.ls.GetGlobal("NodeLibrary").(*lua.LFunction)
	fn := c.ls.GetGlobal("node_library")
	log.Printf("fn=%#v", fn)
	c.showTop()

	lv := c.ls.Get(-1) // get the value at the top of the stack
	if tbl, ok := lv.(*lua.LTable); ok {
		// lv is LTable
		fmt.Println(c.ls.ObjLen(tbl))
	}

	return result
}
