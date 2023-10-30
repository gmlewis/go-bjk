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
	// log.Printf("extrudeWithCaps: amount=%v", amount)
	// Currently, only a single face is extruded.
	face := checkMesh(ls, 3)
	// log.Printf("extrudeWithCaps: face=%v", face)
	// Only the first 3 points in the face are used to calculate its normal.
	faceNormalVec3 := face.CalcNormal(0).MulScalar(amount)
	// log.Printf("extrudeWithCaps: faceNormalVec3=%v", faceNormalVec3)
	faceNormal := NewMeshFromLine(&Vec3{0, 0, 0}, &faceNormalVec3, 1)
	// log.Printf("extrudeWithCaps: faceNormal=%v", faceNormal)

	mesh := NewMeshFromExtrudeAlongCurve(faceNormal, face, 0)
	// Because extrude_along_curve does not make a face at the start or end
	// of the curve, we need to move the initial face to the end of the extrusion
	// before we reverse the original face.
	numVerts := len(face.Verts)
	meshVerts := len(mesh.Verts)
	movedFace := make([]int, 0, numVerts)
	for i := 0; i < numVerts; i++ {
		movedFace = append(movedFace, face.Faces[0][i]+meshVerts-numVerts)
	}
	// log.Printf("extrudeWithCaps: numVerts=%v, meshVerts=%v, movedFace=%v", numVerts, meshVerts, movedFace)
	mesh.Faces = append(mesh.Faces, movedFace)

	// face is altered in-place - so reverse the order of its face[0] vertex indices,
	// then merge the new mesh into it.
	// log.Printf("extrudeWithCaps: before face reversal=%v", face)
	slices.Reverse(face.Faces[0])
	// log.Printf("extrudeWithCaps: after face reversal=%v", face)
	face.Merge(mesh)
	// log.Printf("extrudeWithCaps: after merge=%v", face)

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
