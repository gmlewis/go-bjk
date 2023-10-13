package nodes

import lua "github.com/yuin/gopher-lua"

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

const luaVec3TypeName = "Vec3"

func registerVec3Type(L *lua.LState) {
	mt := L.NewTypeMetatable(luaVec3TypeName)
	L.SetGlobal(luaVec3TypeName, mt)
	// static attributes
	L.SetField(mt, "new", L.NewFunction(newVec3))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), vec3Methods))
}

// Constructor
func newVec3(L *lua.LState) int {
	vec3 := &Vec3{
		X: float64(L.CheckNumber(1)),
		Y: float64(L.CheckNumber(2)),
		Z: float64(L.CheckNumber(3)),
	}
	ud := L.NewUserData()
	ud.Value = vec3
	L.SetMetatable(ud, L.GetTypeMetatable(luaVec3TypeName))
	L.Push(ud)
	return 1
}

// Checks whether the first lua argument is a *LUserData with *Vec3 and returns this *Vec3.
func checkVec3(L *lua.LState) *Vec3 {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Vec3); ok {
		return v
	}
	L.ArgError(1, "vec3 expected")
	return nil
}

var vec3Methods = map[string]lua.LGFunction{
	"x": vec3GetSetX,
	"y": vec3GetSetY,
	"z": vec3GetSetZ,
}

// Getter and setter for the Vec3#x
func vec3GetSetX(L *lua.LState) int {
	p := checkVec3(L)
	if L.GetTop() == 2 {
		p.X = float64(L.CheckNumber(2))
		return 0
	}
	L.Push(lua.LNumber(p.X))
	return 1
}

// Getter and setter for the Vec3#y
func vec3GetSetY(L *lua.LState) int {
	p := checkVec3(L)
	if L.GetTop() == 2 {
		p.Y = float64(L.CheckNumber(2))
		return 0
	}
	L.Push(lua.LNumber(p.Y))
	return 1
}

// Getter and setter for the Vec3#z
func vec3GetSetZ(L *lua.LState) int {
	p := checkVec3(L)
	if L.GetTop() == 2 {
		p.Z = float64(L.CheckNumber(2))
		return 0
	}
	L.Push(lua.LNumber(p.Z))
	return 1
}
