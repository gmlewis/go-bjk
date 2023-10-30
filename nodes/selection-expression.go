package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

// SelectionExpression represents a selection of vertices, edges, or faces.
type SelectionExpression string

const luaSelectionExpressionTypeName = "SelectionExpression"

var selectionExpressionFuncs = map[string]lua.LGFunction{
	"new": newSelectionExpression, // SelectionExpression.new("*")  // or (1) or (1, 2, 3)
}

func registerSelectionExpressionType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaSelectionExpressionTypeName)
	ls.SetGlobal(luaSelectionExpressionTypeName, mt)
	for name, fn := range selectionExpressionFuncs {
		mt.RawSetString(name, ls.NewFunction(fn))
	}
}

// constructor
func newSelectionExpression(ls *lua.LState) int {
	se := SelectionExpression(ls.CheckString(1))
	ud := ls.NewUserData()
	ud.Value = se
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaSelectionExpressionTypeName))
	ls.Push(ud)
	return 1
}
