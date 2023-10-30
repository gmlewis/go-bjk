package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

const luaOpsTypeName = "Ops"

var opsFuncs = map[string]lua.LGFunction{
	"extrude_along_curve": extrudeAlongCurve,
}

func registerOpsType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaOpsTypeName)
	ls.SetGlobal(luaOpsTypeName, mt)
	for name, fn := range opsFuncs {
		mt.RawSetString(name, ls.NewFunction(fn))
	}
}

func extrudeAlongCurve(ls *lua.LState) int {
	backbone := checkMesh(ls, 1)
	crossSection := checkMesh(ls, 2)
	flip := int(ls.CheckNumber(3))

	mesh := NewMeshFromExtrudeAlongCurve(backbone, crossSection, flip)

	ud := ls.NewUserData()
	ud.Value = mesh
	ls.Push(ud)
	return 1
}
