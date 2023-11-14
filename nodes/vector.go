package nodes

import (
	"fmt"
	"log"
	"math"

	"github.com/gmlewis/go3d/float64/mat4"
	"github.com/gmlewis/go3d/float64/quaternion"
	"github.com/gmlewis/go3d/float64/vec3"
	"github.com/gmlewis/go3d/float64/vec4"
	lua "github.com/yuin/gopher-lua"
)

const (
	// Epsilon is a small number.
	Epsilon = 1e-5
)

// Vec3 represents a point in 3D space.
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3) String() string {
	return fmt.Sprintf("%0.5f %0.5f %0.5f", v.X, v.Y, v.Z) // better hashing without surrounding {}
}

// NewVec3 returns a new Vec3.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// AboutEq returns true if a is within Epsilon of b.
func AboutEq(a, b float64) bool { return math.Abs(a-b) < Epsilon }

// AboutZero returns true if entire vector is within Epsilon of (0,0,0)
func (v *Vec3) AboutZero() bool {
	return AboutEq(v.X, 0) && AboutEq(v.Y, 0) && AboutEq(v.Z, 0)
}

// AboutEq returns true if entire vector is within Epsilon of otherVec.
func (v *Vec3) AboutEq(otherVec Vec3) bool {
	return AboutEq(v.X, otherVec.X) && AboutEq(v.Y, otherVec.Y) && AboutEq(v.Z, otherVec.Z)
}

// Negated returns the opposite (negated) vector of v.
func (v Vec3) Negated() Vec3 {
	return Vec3{X: -v.X, Y: -v.Y, Z: -v.Z}
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

// Normalized returns a new normalized Vec3.
func (v Vec3) Normalized() Vec3 {
	vp := &v
	vp.Normalize()
	return v
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

// Vec3Dot performs the dot product of v1 x v2 and returns a new vector.
func Vec3Dot(v1, v2 Vec3) float64 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

// Rotation calculates the rotation between two normalized vectors.
// From: https://physicsforgames.blogspot.com/2010/03/quaternion-tricks.html
func Rotation(from, to Vec3) quaternion.T {
	h := Vec3Add(from, to)
	h.Normalize()
	result := quaternion.T{}
	result[3] = Vec3Dot(from, h)
	result[0] = from.Y*h.X - from.Z*h.Y
	result[1] = from.Z*h.X - from.X*h.Z
	result[2] = from.X*h.Y - from.Y*h.X
	return result
}

// Cross returns the cross product of v1 x v2 and returns a new vector.
func (v1 Vec3) Cross(v2 Vec3) Vec3 { return Vec3Cross(v1, v2) }

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

// constructor
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

// GenXform represents a rotation (defined by the normal and tangent vectors) about the origin,
// then translating it by tr into place.
// func GenXform(normal, tangent, tr Vec3) *Xform {
func GenXform(normal, tangent, tr Vec3) *mat4.T {
	normal.Normalize()
	tangent.Normalize()
	cotangent := normal.Cross(tangent)
	cotangent.Normalize()
	// log.Printf("normal=%v, tangent=%v, cotangent=%v", normal, tangent, cotangent)

	xAxis := vec3.T{cotangent.X, cotangent.Y, cotangent.Z}
	yAxis := vec3.T{normal.X, normal.Y, normal.Z}
	zAxis := vec3.T{tangent.X, tangent.Y, tangent.Z}
	rot := quaternion.FromRotationAxes(xAxis, yAxis, zAxis)

	// Convert to 4x4 transformation matrix:
	x2 := 2 * rot[0]
	y2 := 2 * rot[1]
	z2 := 2 * rot[2]
	xx := rot[0] * x2
	xy := rot[0] * y2
	xz := rot[0] * z2
	yy := rot[1] * y2
	yz := rot[1] * z2
	zz := rot[2] * z2
	wx := rot[3] * x2
	wy := rot[3] * y2
	wz := rot[3] * z2

	return &mat4.T{ // note column order due to MulVec4 ordering.
		vec4.T{1.0 - (yy + zz), xy + wz, xz - wy, 0},
		vec4.T{xy - wz, 1.0 - (xx + zz), yz + wx, 0},
		vec4.T{xz + wy, yz - wx, 1.0 - (xx + yy), 0},
		vec4.T{tr.X, tr.Y, tr.Z, 1},
	}
}

// Xform applies the 4x4 transformation matrix to the provided vector and
// returns the result.
func (v Vec3) Xform(xform *mat4.T) Vec3 {
	result := xform.MulVec4(&vec4.T{v.X, v.Y, v.Z, 1})
	// log.Printf("Xform: v=%v, result=%v", v, result)
	return Vec3{X: result[0], Y: result[1], Z: result[2]}
}

// Useful for debugging lua:
//     print("GML: VectorMath: op=", inputs.op, ", vec_a=", inputs.vec_a, ", vec_b=", inputs.vec_b)
//     local mt = getmetatable(inputs.vec_a)
//     for k, v in pairs(mt) do
//         print("k=", k, ", v=", v)
//     end
