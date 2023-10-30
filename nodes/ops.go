package nodes

import (
	"slices"

	lua "github.com/yuin/gopher-lua"
)

const luaOpsTypeName = "Ops"

var opsFuncs = map[string]lua.LGFunction{
	"extrude_along_curve": extrudeAlongCurve,
	"extrude_with_caps":   extrudeWithCaps,
	"merge":               mergeMeshes,
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
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

func extrudeWithCaps(ls *lua.LState) int {
	// Currently, the SelectionExpression (first arg) is assumed to be '*'.
	amount := float64(ls.CheckNumber(2))
	// Currently, only a single face is extruded.
	face := checkMesh(ls, 3)
	// Only the first 3 points in the face are used to calculate its normal.
	faceNormalVec3 := face.CalcNormal().MulScalar(amount)
	faceNormal := NewMeshFromLine(&Vec3{0, 0, 0}, &faceNormalVec3, 1)

	mesh := NewMeshFromExtrudeAlongCurve(faceNormal, face, 0)

	// face is altered in-place - so reverse the order of its face[0] vertex indices,
	// then merge the new mesh into it.
	slices.Reverse(face.Faces[0])
	face.Merge(mesh)

	return 0
}

// mergeMeshes merges src into dst for Ops.merge(dst, src).
// It returns nothing on the stack.
func mergeMeshes(ls *lua.LState) int {
	dst := checkMesh(ls, 1)
	src := checkMesh(ls, 2)

	dst.Merge(src)
	return 0
}
