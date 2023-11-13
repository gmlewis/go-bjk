package nodes

import (
	"log"
	"sort"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/exp/maps"
)

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	Verts    []Vec3
	Normals  []Vec3  // optional - per-vert normals
	Tangents []Vec3  // optional - per-vert tangents
	Faces    []FaceT // optional - when used, Normals and Tangents are unused.
}

// FaceT represents a face and is a slice of vertex indices.
type FaceT []int

const luaMeshTypeName = "Mesh"

var meshFuncs = map[string]lua.LGFunction{
	"clone": meshClone, // mesh_a:clone()
}

func registerMeshType(ls *lua.LState) {
	mt := ls.NewTypeMetatable(luaMeshTypeName)
	ls.SetGlobal(luaMeshTypeName, mt)
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), meshFuncs))
	// for name, fn := range meshFuncs {
	// 	mt.RawSetString(name, ls.NewFunction(fn))
	// }
}

func meshClone(ls *lua.LState) int {
	orig := checkMesh(ls, 1)
	m := &Mesh{
		Verts:    make([]Vec3, 0, len(orig.Verts)),
		Normals:  make([]Vec3, 0, len(orig.Normals)),
		Tangents: make([]Vec3, 0, len(orig.Tangents)),
		Faces:    make([]FaceT, 0, len(orig.Faces)),
	}

	m.Verts = append(m.Verts, orig.Verts...)
	m.Normals = append(m.Normals, orig.Normals...)
	m.Tangents = append(m.Tangents, orig.Tangents...)
	for _, v := range orig.Faces {
		m.Faces = append(m.Faces, append(FaceT{}, v...))
	}

	ud := ls.NewUserData()
	ud.Value = m
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

// NewPolygonFromPoints creates a new mesh from points.
func NewPolygonFromPoints(pts []Vec3) *Mesh {
	m := &Mesh{Verts: pts, Faces: []FaceT{make(FaceT, 0, len(pts))}}
	for i := 0; i < len(pts); i++ {
		m.Faces[0] = append(m.Faces[0], i)
	}
	return m
}

// NewMeshFromPolygons creates a new mesh from points.
func NewMeshFromPolygons(verts []Vec3, faces []FaceT) *Mesh {
	// log.Printf("NewMeshFromPolygons: %v verts, %v faces", len(verts), len(faces))
	return &Mesh{Verts: verts, Faces: faces}
}

// NewMeshFromLineWithNormals creates a new mesh from points, normals, and tangents.
func NewMeshFromLineWithNormals(points, normals, tangents []Vec3) *Mesh {
	// log.Printf("NewMeshFromLineWithNormals: %v points, %v normals, %v tangents", len(points), len(normals), len(tangents))
	return &Mesh{
		Verts:    points,
		Normals:  normals,
		Tangents: tangents,
	}
}

// NewMeshFromLine creates a new mesh from two points, divided into numSegs.
func NewMeshFromLine(v1, v2 *Vec3, numSegs int) *Mesh {
	// log.Printf("NewMeshFromLine: 2 points, %v segments", numSegs)
	m := &Mesh{
		Verts: make([]Vec3, 0, numSegs+1),
	}
	lerp := func(val1, val2 float64, i int) float64 {
		t := float64(i) / float64(numSegs)
		return (val2-val1)*t + val1
	}
	for i := 0; i < numSegs; i++ {
		v := Vec3{
			X: lerp(v1.X, v2.X, i),
			Y: lerp(v1.Y, v2.Y, i),
			Z: lerp(v1.Z, v2.Z, i),
		}
		m.Verts = append(m.Verts, v)
	}
	m.Verts = append(m.Verts, *v2)
	return m
}

// NewMeshFromExtrudeAlongCurve creates a new mesh by extruding the crossSection along the backbone.
// Note that extrude along curve in Blackjack does not make a face at the start or end of the curve.
func NewMeshFromExtrudeAlongCurve(backbone, crossSection *Mesh, flip int) *Mesh {
	if len(backbone.Verts) == 0 || len(crossSection.Verts) == 0 || len(backbone.Normals) < len(backbone.Verts) {
		log.Printf("NewMeshFromExtrudeAlongCurve not enough verts(%v/%v) or normals(%v) to extrude",
			len(backbone.Verts), len(crossSection.Verts), len(backbone.Normals))
		return &Mesh{}
	}

	numVerts := len(crossSection.Verts)
	m := &Mesh{
		Verts: make([]Vec3, 0, numVerts*len(backbone.Verts)),
		Faces: make([]FaceT, 0, numVerts*(len(backbone.Verts)-1)),
	}

	if len(backbone.Tangents) < len(backbone.Verts) {
		backbone.generateTangents()
	}

	// For each segment, add numVerts to the mesh, rotated and translated into place, and create new faces
	// that connect to the last set of numVerts.
	for bvi := 0; bvi < len(backbone.Verts); bvi++ {

		normal, tangent, bvert := backbone.Normals[bvi], backbone.Tangents[bvi], backbone.Verts[bvi]
		xform := GenXform(normal, tangent, bvert)
		vIdx := len(m.Verts)
		// log.Printf("nmfeac: bvi=%v, normal=%v, tangent=%v, bvert=%v, xform=%v, vIdx=%v", bvi, backbone.Normals[bvi], backbone.Tangents[bvi], bvert, xform, vIdx)
		for i, v := range crossSection.Verts {
			// m.Verts = append(m.Verts, xform.Do(v))
			m.Verts = append(m.Verts, v.Xform(xform))
			// log.Printf("verts[%v]=%v", len(m.Verts)-1, m.Verts[len(m.Verts)-1])
			if bvi == 0 {
				continue
			}

			// create a new quad for each extruded crossSection vertex
			m.Faces = append(m.Faces, FaceT{
				vIdx + i - numVerts,
				vIdx + i,
				vIdx + ((i + 1) % numVerts),
				vIdx + ((i + 1) % numVerts) - numVerts,
			})
			// log.Printf("face[%v]=%+v", len(m.Faces)-1, m.Faces[len(m.Faces)-1])
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

// CalcFaceNormal calculates the normal of a face.
// Note that concave polygons are problematic if the wrong vertices are chosen.
//
// One way to solve this "correctly" would be to see if any edge used to create the normal
// lies outside the polygon, but this is computationally and algorithmically complex.
// This implementation uses a simple heuristic using a voting mechanism.
func (m *Mesh) CalcFaceNormal(face FaceT) Vec3 {
	if len(m.Verts) < 3 || len(face) < 3 {
		log.Fatalf("CalcNormalAndTangent: want >=3 points >=1 face, got %#v", *m)
	}

	votes := map[Vec3]int{}

	numVerts := len(face)
	for i, vIdx := range face {
		va := m.Verts[vIdx]
		vb := m.Verts[face[(i+1)%numVerts]]
		vc := m.Verts[face[(i-1+numVerts)%numVerts]]
		// log.Printf("i=%v, va=%v, vb=%v, vc=%v", i, va, vb, vc)
		n := (vb.Sub(va)).Cross(vc.Sub(va)).Normalized()
		// log.Printf("(vb-va)x(vc-va)=%v", n)
		votes[n]++
	}

	keys := maps.Keys(votes)
	sort.Slice(keys, func(i, j int) bool {
		return votes[keys[i]] > votes[keys[j]]
	})

	return keys[0]
}
