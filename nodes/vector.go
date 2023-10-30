package nodes

import (
	"fmt"
	"log"
	"math"

	"github.com/gmlewis/go3d/float64/mat4"
	"github.com/gmlewis/go3d/float64/vec3"
	"github.com/gmlewis/go3d/float64/vec4"
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

// Normalize normalizes a vector in-place.
func (v *Vec3) Normalize() {
	length := v.Length()
	if length > 0 {
		v.X /= length
		v.Y /= length
		v.Z /= length
	}
}

// Length calculates the length of the vector.
func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
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

// Sub subtracts vector v2 from vector v1 and returns a new vector.
func (v1 Vec3) Sub(v2 Vec3) Vec3 { return Vec3Sub(v1, v2) }

// Vec3Mul multiplies two vectors (element by element) and returns a new vector.
func Vec3Mul(v1, v2 Vec3) Vec3 {
	return Vec3{X: v1.X * v2.X, Y: v1.Y * v2.Y, Z: v1.Z * v2.Z}
}

// Vec3Cross performs the cross product of v1 x v2 and returns a new vector.
func Vec3Cross(v1, v2 Vec3) Vec3 {
	return Vec3{
		X: v1.Y*v2.Z - v1.Z*v2.Y,
		Y: v1.Z*v2.X - v1.X*v2.Z,
		Z: v1.X*v2.Y - v1.Y*v2.X,
	}
}

// GetRotXYZ gets the x, y, and z rotations for a vector.
func (v Vec3) GetRotXYZ() (rx, ry, rz float64) {
	rx = math.Atan2(v.Z, v.Y)
	ry = math.Atan2(v.Z, v.X)
	rz = math.Atan2(v.Y, v.X)
	return rx, ry, rz
}

// Cross returns the cross product of v1 x v2 and returns a new vector.
func (v1 Vec3) Cross(v2 *Vec3) Vec3 { return Vec3Cross(v1, *v2) }

const luaVec3TypeName = "Vec3"

var vec3Funcs = map[string]lua.LGFunction{
	"new":        newVec3,
	"__index":    vec3Index,
	"__add":      vec3Add,
	"__sub":      vec3Sub,
	"__mul":      vec3Mul,
	"__tostring": vec3String,
}

func registerVec3Type(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaVec3TypeName)
	ls.SetGlobal(luaVec3TypeName, mt)
	for name, fn := range vec3Funcs {
		mt.RawSetString(name, ls.NewFunction(fn))
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

func vec3Index(ls *lua.LState) int {
	p := checkVec3(ls, 1)
	if p == nil {
		return 0
	}

	key := ls.CheckString(2)
	switch key {
	case "x":
		ls.Push(lua.LNumber(p.X))
	case "y":
		ls.Push(lua.LNumber(p.Y))
	case "z":
		ls.Push(lua.LNumber(p.Z))
	default:
		log.Fatalf("vec3Index - unexpected key '%v'", key)
	}
	return 1
}

// checkVec3 checks whether the first lua argument is a *LUserData with *Vec3 and returns this *Vec3.
func checkVec3(ls *lua.LState, index int) *Vec3 {
	// log.Printf("checkVec3: Get(%v): (%v,%v)", index, ls.Get(index).String(), ls.Get(index).Type())
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
	// log.Printf("vec3GetSetX: returning X=%v to lua", p.X)
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
	// log.Printf("vec3GetSetY: returning Y=%v to lua", p.Y)
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
	// log.Printf("vec3GetSetZ: returning Z=%v to lua", p.Z)
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
	lhs, rhs := ls.Get(1).Type(), ls.Get(2).Type()
	var p1, p2 *Vec3
	switch {
	case lhs == lua.LTUserData && rhs == lua.LTNumber:
		p1 = checkVec3(ls, 1)
		v := float64(ls.CheckNumber(2))
		p2 = &Vec3{X: v, Y: v, Z: v}
	case lhs == lua.LTNumber && rhs == lua.LTUserData:
		v := float64(ls.CheckNumber(1))
		p1 = &Vec3{X: v, Y: v, Z: v}
		p2 = checkVec3(ls, 2)
	case lhs == lua.LTUserData && rhs == lua.LTUserData:
		p1 = checkVec3(ls, 1)
		p2 = checkVec3(ls, 2)
	default:
		log.Fatalf("unhandled vec3op2 between lhs=%q and rhs=%q", lhs, rhs)
	}

	// log.Printf("vec3op2: p1=%v, p2=%v", p1, p2)

	result := opFn(*p1, *p2)
	ud := ls.NewUserData()
	ud.Value = &result
	// log.Printf("vec3op2 result = %v", result)
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaVec3TypeName))
	ls.Push(ud)
	return 1
}

func vec3Add(ls *lua.LState) int {
	// log.Printf("GML: RUNNING vec3Add from Lua")
	return vec3op2(ls, Vec3Add)
}

func vec3Sub(ls *lua.LState) int {
	// log.Printf("GML: RUNNING vec3Sub from Lua")
	return vec3op2(ls, Vec3Sub)
}

func vec3Mul(ls *lua.LState) int {
	// log.Printf("GML: RUNNING vec3Mul from Lua")
	return vec3op2(ls, Vec3Mul)
}

func vec3Cross(ls *lua.LState) int {
	// log.Printf("GML: RUNNING vec3Cross from Lua")
	return vec3op2(ls, Vec3Cross)
}

// Useful for debugging lua:
//     print("GML: VectorMath: op=", inputs.op, ", vec_a=", inputs.vec_a, ", vec_b=", inputs.vec_b)
//     local mt = getmetatable(inputs.vec_a)
//     for k, v in pairs(mt) do
//         print("k=", k, ", v=", v)
//     end

// GenXform generates a 4x4 transformation matrix by rotating by rx, ry, rz about the origin,
// then translating it by tr into place.
func GenXform(rx, ry, rz float64, tr Vec3) *mat4.T {
	xrot := mat4.T{}
	xrot.AssignXRotation(rx)
	yrot := mat4.T{}
	yrot.AssignXRotation(ry)
	zrot := mat4.T{}
	zrot.AssignXRotation(rz)
	xyzt := mat4.Ident
	xyzt.Translate(&vec3.T{tr.X, tr.Y, tr.Z})
	xfrm := mat4.T{}
	xfrm.AssignMul(&xrot, &yrot)
	xfrm2 := mat4.T{}
	xfrm2.AssignMul(&xfrm, &zrot)
	xfrm3 := mat4.T{}
	xfrm3.AssignMul(&xfrm2, &xyzt)
	return &xfrm3
}

// Xform applies the 4x4 transformation matrix to the provided vector and
// returns the result.
func (v Vec3) Xform(xform *mat4.T) Vec3 {
	result := xform.MulVec4(&vec4.T{v.X, v.Y, v.Z, 1})
	return Vec3{X: result[0], Y: result[1], Z: result[2]}
}
