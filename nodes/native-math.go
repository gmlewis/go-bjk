package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

const luaNativeMathTypeName = "NativeMath"

var nativeMathFuncs = map[string]lua.LGFunction{
	"cross":              vec3Cross,
	"rotate_around_axis": rotateAroundAxis,
}

func registerNativeMathType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaNativeMathTypeName)
	ls.SetGlobal(luaNativeMathTypeName, mt)
	for name, fn := range nativeMathFuncs {
		mt.RawSetString(name, ls.NewFunction(fn))
	}
}
