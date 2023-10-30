package nodes

import (
	lua "github.com/yuin/gopher-lua"
)

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	Verts    []Vec3
	Normals  []Vec3  // optional - per-vert normals
	Tangents []Vec3  // optional - per-vert tangents
	Faces    [][]int // optional
}

// NewMeshFromPolygons creates a new mesh from points.
func NewMeshFromPolygons(verts []Vec3, faces [][]int) *Mesh {
	return &Mesh{Verts: verts, Faces: faces}
}

// NewMeshFromLineWithNormals creates a new mesh from points, normals, and tangents.
func NewMeshFromLineWithNormals(points, normals, tangents []Vec3) *Mesh {
	return &Mesh{
		Verts:    points,
		Normals:  normals,
		Tangents: tangents,
	}
}

// NewMeshFromExtrudeAlongCurve creates a new mesh by extruding the crossSection along the backbone.
func NewMeshFromExtrudeAlongCurve(backbone, crossSection *Mesh, flip int) *Mesh {
	if len(backbone.Verts) == 0 || len(crossSection.Verts) == 0 {
		return &Mesh{}
	}

	numVerts := len(crossSection.Verts)
	m := &Mesh{
		Verts: make([]Vec3, 0, numVerts*len(backbone.Verts)),
		Faces: make([][]int, 0, numVerts*(len(backbone.Verts)-1)),
	}
	startPos := backbone.Verts[0]
	if len(backbone.Tangents) < len(backbone.Verts) {
		backbone.generateTangents()
	}

	baseRotX, baseRotY, baseRotZ := backbone.Tangents[0].GetRotXYZ()

	// First, make a copy of the crossSection verts positioned in-place at the start of the backbone.
	for _, v := range crossSection.Verts {
		m.Verts = append(m.Verts, v.Add(startPos))
	}

	// For each segment, add numVerts to the mesh, rotated and translated into place, and create new faces
	// that connect to the last set of numVerts.
	for bvi := 1; bvi < len(backbone.Verts); bvi++ {
		rotX, rotY, rotZ := backbone.Tangents[bvi].GetRotXYZ()
		rx, ry, rz := rotX-baseRotX, rotY-baseRotY, rotZ-baseRotZ

		bvert := backbone.Verts[bvi]
		xform := GenXform(rx, ry, rz, bvert)
		vIdx := len(m.Verts)
		for i, v := range crossSection.Verts {
			m.Verts = append(m.Verts, v.Xform(xform))
			// create a new quad for each extruded crossSection vertex
			m.Faces = append(m.Faces, []int{
				vIdx + i - numVerts,
				vIdx + i,
				vIdx + ((i + 1) % numVerts),
				vIdx + ((i + 1) % numVerts) - numVerts,
			})
		}
	}

	return m
}

// checkMesh checks whether the first lua argument is a *LUserData with *Mesh and returns this *Mesh.
func checkMesh(ls *lua.LState, index int) *Mesh {
	// log.Printf("checkMesh: Get(%v): (%v,%v)", index, ls.Get(index).String(), ls.Get(index).Type())
	ud := ls.CheckUserData(index)
	if v, ok := ud.Value.(*Mesh); ok {
		return v
	}
	ls.ArgError(index, "mesh expected")
	return nil
}

func (m *Mesh) generateTangents() {
	m.Tangents = make([]Vec3, 0, len(m.Verts))
	for i := 1; i < len(m.Verts); i++ {
		v := Vec3{
			X: m.Verts[i].X - m.Verts[i-1].X,
			Y: m.Verts[i].Y - m.Verts[i-1].Y,
			Z: m.Verts[i].Z - m.Verts[i-1].Z,
		}
		v.Normalize()
		m.Tangents = append(m.Tangents, v)
	}
	// last tangent doesn't matter, but tangents must be same length as verts.  Copy last one.
	m.Tangents = append(m.Tangents, m.Tangents[len(m.Tangents)-1])
}
