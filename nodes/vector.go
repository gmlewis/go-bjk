package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

const luaVec3TypeName = "Vec3"

func registerVec3Type(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaVec3TypeName)
	ls.SetGlobal(luaVec3TypeName, mt)
	// static attributes
	ls.SetField(mt, "new", ls.NewFunction(newVec3))
	// methods
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), vec3Methods))
}

// Constructor
func newVec3(ls *lua.LState) int {
	vec3 := &Vec3{
		X: float64(ls.CheckNumber(1)),
		Y: float64(ls.CheckNumber(2)),
		Z: float64(ls.CheckNumber(3)),
	}
	ud := ls.NewUserData()
	ud.Value = vec3
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaVec3TypeName))
	ls.Push(ud)
	return 1
}

// Checks whether the first lua argument is a *LUserData with *Vec3 and returns this *Vec3.
func checkVec3(ls *lua.LState) *Vec3 {
	ud := ls.CheckUserData(1)
	if v, ok := ud.Value.(*Vec3); ok {
		return v
	}
	ls.ArgError(1, "vec3 expected")
	return nil
}

var vec3Methods = map[string]lua.LGFunction{
	"x": vec3GetSetX,
	"y": vec3GetSetY,
	"z": vec3GetSetZ,
}

// Getter and setter for the Vec3#x
func vec3GetSetX(ls *lua.LState) int {
	p := checkVec3(ls)
	if ls.GetTop() == 2 {
		p.X = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.X))
	return 1
}

// Getter and setter for the Vec3#y
func vec3GetSetY(ls *lua.LState) int {
	p := checkVec3(ls)
	if ls.GetTop() == 2 {
		p.Y = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.Y))
	return 1
}

// Getter and setter for the Vec3#z
func vec3GetSetZ(ls *lua.LState) int {
	p := checkVec3(ls)
	if ls.GetTop() == 2 {
		p.Z = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.Z))
	return 1
}
