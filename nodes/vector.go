package nodes

import (
	"fmt"
	"log"

	lua "github.com/yuin/gopher-lua"
)

// Vec3 represents a point in 3D space.
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

// NewVec3 returns a new Vec3.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// MulScalar multiplies a Vec3 by a scalar value and returns a new vector.
// v is not altered.
func (v Vec3) MulScalar(f float64) Vec3 {
	return Vec3{X: v.X * f, Y: v.Y * f, Z: v.Z * f}
}

// Vec3Add adds two vectors and returns a new vector.
func Vec3Add(v1, v2 Vec3) Vec3 { return Vec3{X: v1.X + v2.X, Y: v1.Y + v2.Y, Z: v1.Z + v2.Z} }

// Add adds two vectors and returns a new vector.
func (v1 Vec3) Add(v2 Vec3) Vec3 { return Vec3Add(v1, v2) }

// Vec3Sub subtracts vector v2 from vector v1 and returns a new vector.
func Vec3Sub(v1, v2 Vec3) Vec3 {
	return Vec3{X: v1.X - v2.X, Y: v1.Y - v2.Y, Z: v1.Z - v2.Z}
}

// Vec3Mul multiplies two vectors (element by element) and returns a new vector.
func Vec3Mul(v1, v2 Vec3) Vec3 {
	return Vec3{X: v1.X * v2.X, Y: v1.Y * v2.Y, Z: v1.Z * v2.Z}
}

const luaVec3TypeName = "Vec3"

var vec3Methods = map[string]lua.LGFunction{
	"x":          vec3GetSetX,
	"y":          vec3GetSetY,
	"z":          vec3GetSetZ,
	"__add":      vec3Add,
	"__sub":      vec3Sub,
	"__mul":      vec3Mul,
	"__tostring": vec3String,
}

func registerVec3Type(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaVec3TypeName)
	ls.SetGlobal(luaVec3TypeName, mt)
	// static attributes
	ls.SetField(mt, "new", ls.NewFunction(newVec3))
	// methods
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), vec3Methods))
	// not sure why this is necessary
	for name, fn := range vec3Methods {
		ls.SetField(mt, name, ls.NewFunction(fn))
	}
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
func checkVec3(ls *lua.LState, index int) *Vec3 {
	ud := ls.CheckUserData(index)
	if v, ok := ud.Value.(*Vec3); ok {
		return v
	}
	ls.ArgError(index, "vec3 expected")
	return nil
}

// Getter and setter for the Vec3#x
func vec3GetSetX(ls *lua.LState) int {
	p := checkVec3(ls, 1)
	if p == nil {
		return 0
	}
	if ls.GetTop() == 2 {
		p.X = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.X))
	return 1
}

// Getter and setter for the Vec3#y
func vec3GetSetY(ls *lua.LState) int {
	p := checkVec3(ls, 1)
	if p == nil {
		return 0
	}
	if ls.GetTop() == 2 {
		p.Y = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.Y))
	return 1
}

// Getter and setter for the Vec3#z
func vec3GetSetZ(ls *lua.LState) int {
	p := checkVec3(ls, 1)
	if p == nil {
		return 0
	}
	if ls.GetTop() == 2 {
		p.Z = float64(ls.CheckNumber(2))
		return 0
	}
	ls.Push(lua.LNumber(p.Z))
	return 1
}

func vec3String(ls *lua.LState) int {
	p := checkVec3(ls, 1)
	if p == nil {
		return 0
	}
	ls.Push(lua.LString(fmt.Sprintf("vector(%v,%v,%v)", p.X, p.Y, p.Z)))
	return 1
}

func vec3op2(ls *lua.LState, opFn func(v1, v2 Vec3) Vec3) int {
	p1 := checkVec3(ls, 1)
	if p1 == nil {
		return 0
	}
	p2 := checkVec3(ls, 2)
	if p2 == nil {
		return 0
	}

	result := opFn(*p1, *p2)
	ud := ls.NewUserData()
	ud.Value = &result
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaVec3TypeName))
	ls.Push(ud)
	return 1
}

func vec3Add(ls *lua.LState) int {
	log.Printf("GML: RUNNING vec3Add from Lua")
	return vec3op2(ls, Vec3Add)
}

func vec3Sub(ls *lua.LState) int {
	log.Printf("GML: RUNNING vec3Sub from Lua")
	return vec3op2(ls, Vec3Sub)
}

func vec3Mul(ls *lua.LState) int {
	log.Printf("GML: RUNNING vec3Mul from Lua")
	return vec3op2(ls, Vec3Mul)
}

// Useful for debugging lua:
//     print("GML: VectorMath: op=", inputs.op, ", vec_a=", inputs.vec_a, ", vec_b=", inputs.vec_b)
//     local mt = getmetatable(inputs.vec_a)
//     for k, v in pairs(mt) do
//         print("k=", k, ", v=", v)
//     end
