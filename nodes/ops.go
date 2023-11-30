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
	"lerp_along_curve":    lerpAlongCurve,
	"merge":               mergeMeshes,
}

func registerOpsType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaOpsTypeName)
	ls.SetGlobal(luaOpsTypeName, mt)
	for name, fn := range opsFuncs {
		mt.RawSetString(name, ls.NewFunction(fn))
	}
}

func lerpAlongCurve(ls *lua.LState) int {
	t := ls.CheckNumber(1)
	curve := checkMesh(ls, 2)

	vec3 := curve.LerpAlongCurve(float64(t))

	ud := newVec3LValue(ls, vec3)
	ls.Push(ud)
	return 1
}

func extrudeAlongCurve(ls *lua.LState) int {
	backbone := checkMesh(ls, 1)
	// log.Printf("extrudeAlongCurve: backbone=%v", backbone)
	crossSection := checkMesh(ls, 2)
	// log.Printf("extrudeAlongCurve: crossSection=%v", crossSection)
	flip := int(ls.CheckNumber(3))

	mesh := NewMeshFromExtrudeAlongCurve(backbone, crossSection, flip)
	// log.Printf("extrudeAlongCurve: mesh=%v", mesh)

	ud := ls.NewUserData()
	ud.Value = mesh
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

func extrudeWithCaps(ls *lua.LState) int {
	// Currently, the SelectionExpression (first arg) is assumed to be '*'.
	amount := float64(ls.CheckNumber(2))
	// log.Printf("\n\nextrudeWithCaps: amount=%v", amount)

	faceMesh := checkMesh(ls, 3)
	// log.Printf("extrudeWithCaps: BEFORE: faceMesh=%v", faceMesh)

	var newFaces []FaceT
	for faceIdx, face := range faceMesh.Faces {
		if len(face) < 3 {
			log.Printf("extrudeWithCaps: Attempted to extrude a face[%v]=%+v with only %v vertices. Skipping.", faceIdx, face, len(face))
			continue
		}

		extrusionNormal := faceMesh.CalcFaceNormal(faceMesh.Faces[faceIdx])
		extrudeVec := extrusionNormal.MulScalar(amount)
		// log.Printf("face[%v]: extrudeVec=%v", faceIdx, extrudeVec)

		// For this face, make another copy of all its vertices at the extruded distance.
		numVerts := len(face)
		vIdx := len(faceMesh.Verts)
		extrudedFace := make(FaceT, 0, numVerts)
		for i, vertIdx := range face {
			// faceMesh.Verts = append(faceMesh.Verts, faceMesh.Verts[vertIdx].Add(extrudeVec))
			addedVertIdx := faceMesh.AddVert(faceMesh.Verts[vertIdx].Add(extrudeVec))
			if addedVertIdx != VertIndexT(vIdx+i) {
				log.Fatalf("extrudeWithCaps: programming error: addedVertIdx(%v) != vIdx(%v)+i(%v)", addedVertIdx, vIdx, i)
			}

			newFaces = append(newFaces, FaceT{
				VertIndexT(vIdx + i),
				VertIndexT(vIdx + i - numVerts),
				VertIndexT(vIdx + ((i + 1) % numVerts) - numVerts),
				VertIndexT(vIdx + ((i + 1) % numVerts)),
			})
			extrudedFace = append(extrudedFace, VertIndexT(vIdx+i))
		}

		// Copy the initial face to the end of the extrusion and make new quads
		// before we reverse the original face.
		newFaces = append(newFaces, extrudedFace)

		// face is altered in-place - so reverse the order of its face[faceIdx] vertex indices.
		slices.Reverse(faceMesh.Faces[faceIdx])
	}

	faceMesh.Faces = append(faceMesh.Faces, newFaces...)
	// log.Printf("extrudeWithCaps: AFTER: faceMesh=%v", faceMesh)

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
