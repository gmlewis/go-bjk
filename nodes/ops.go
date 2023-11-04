package nodes

import (
	"log"
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
	log.Printf("extrudeWithCaps: amount=%v", amount)
	// Currently, only a single face is extruded.
	faceMesh := checkMesh(ls, 3)
	log.Printf("extrudeWithCaps: faceMesh=%v", faceMesh)
	// Only the first 3 points in the face are used to calculate its normal and tangent.
	segmentNormal, segmentTangent := faceMesh.CalcNormalAndTangent(0)
	log.Printf("extrudeWithCaps: segmentNormal=%v, segmentTangent=%v", segmentNormal, segmentTangent)
	startPos := Vec3{0, 0, 0}
	endPos := segmentNormal.MulScalar(amount)
	log.Printf("extrudeWithCaps: endPos=%v", endPos)
	points := []Vec3{startPos, endPos}
	normals := []Vec3{segmentNormal, segmentNormal}
	tangents := []Vec3{segmentTangent, segmentTangent}
	faceNormal := NewMeshFromLineWithNormals(points, normals, tangents)
	log.Printf("extrudeWithCaps: faceNormal=%v", faceNormal)

	mesh := NewMeshFromExtrudeAlongCurve(faceNormal, faceMesh, 0)
	log.Printf("extrudeWithCaps: extruded mesh=%v", mesh)
	// Because extrude_along_curve does not make a face at the start or end
	// of the curve, we need to move the initial face to the end of the extrusion
	// before we reverse the original face.
	numVerts := len(faceMesh.Verts)
	meshVerts := len(mesh.Verts)
	movedFace := make([]int, 0, numVerts)
	for i := 0; i < numVerts; i++ {
		movedFace = append(movedFace, faceMesh.Faces[0][i]+meshVerts-numVerts)
	}
	log.Printf("extrudeWithCaps: numVerts=%v, meshVerts=%v, movedFace=%v", numVerts, meshVerts, movedFace)
	mesh.Faces = append(mesh.Faces, movedFace)

	// face is altered in-place - so reverse the order of its face[0] vertex indices,
	// then merge the new mesh into it.
	log.Printf("extrudeWithCaps: before face reversal=%v", faceMesh)
	slices.Reverse(faceMesh.Faces[0])
	log.Printf("extrudeWithCaps: after face reversal=%v", faceMesh)
	faceMesh.Merge(mesh)
	log.Printf("extrudeWithCaps: after merge=%v", faceMesh)

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
