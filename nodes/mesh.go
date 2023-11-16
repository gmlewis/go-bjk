package nodes

import (
	"fmt"
	"log"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/exp/maps"
)

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	// Do not manually add to Verts. Use AddVert instead.
	Verts       []Vec3
	uniqueVerts map[vertKeyT]VertIndexT

	Normals  []Vec3  // optional - per-vert normals
	Tangents []Vec3  // optional - per-vert tangents
	Faces    []FaceT // optional - when used, Normals and Tangents are unused.
}

// VertIndexT represents a vertex index.
type VertIndexT int

// FaceT represents a face and is a slice of vertex indices.
type FaceT []VertIndexT

// vertKeyT represents a vertex key (or "signature") which uniquely identifies a vertex for easy merging.
type vertKeyT string

// toKey generates a vertex vertKeyT (or "signature") which is a string representation of the vertex.
func (v Vec3) toKey() vertKeyT {
	return vertKeyT(fmt.Sprintf("%0.5f %0.5f %0.5f", v.X, v.Y, v.Z)) // better hashing without surrounding {}
}

// faceKeyT represents a face key (or "signature") which uniquely identifies a face consisting of the same verts.
type faceKeyT string

// toKey generates a faceKeyT (or "signature") which is a string of the sorted vertex indices.
func (f FaceT) toKey() faceKeyT {
	verts := append([]VertIndexT{}, f...)
	sort.Slice(verts, func(i, j int) bool { return verts[i] < verts[j] })
	return faceKeyT(fmt.Sprintf("%v", verts))
}

// faceVertKeyT represents a face with verts key (or "signature") which uniquely identifies a face consisting of the same vert Vec3s.
type faceVertKeyT string

// toVertKey generates a faceVertKeyT (or "signature") which is a string of the sorted vertex Vec3s.
func (is *infoSetT) toVertKey(f FaceT) faceVertKeyT {
	verts := make([]string, 0, len(f))
	for _, vertIdx := range f {
		verts = append(verts, string(is.faceInfo.m.Verts[vertIdx].toKey()))
	}
	sort.Slice(verts, func(i, j int) bool { return verts[i] < verts[j] })
	return faceVertKeyT(strings.Join(verts, ","))
}

// faceIndexT represents a face index and is only used internally.
type faceIndexT int

// AddVert adds a vertex to a mesh (reusing existing vertices if possible) and returns its VertIndexT.
func (m *Mesh) AddVert(v Vec3) VertIndexT {
	key := v.toKey()
	if vertIdx, ok := m.uniqueVerts[key]; ok {
		return vertIdx
	}
	vertIdx := VertIndexT(len(m.Verts))
	m.uniqueVerts[key] = vertIdx
	m.Verts = append(m.Verts, v)
	return vertIdx
}

// AddFace adds a face to a mesh and returns its FaceT.
func (m *Mesh) AddFace(verts []Vec3) FaceT {
	face := make([]VertIndexT, 0, len(verts))
	for _, vert := range verts {
		vertIdx := m.AddVert(vert)
		face = append(face, vertIdx)
	}
	m.Faces = append(m.Faces, face)
	return face
}

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

func newMeshFrom(verts, normals, tangents []Vec3, faces []FaceT) *Mesh {
	m := &Mesh{
		Verts:       make([]Vec3, 0, len(verts)),
		uniqueVerts: map[vertKeyT]VertIndexT{},

		Normals:  make([]Vec3, 0, len(normals)),
		Tangents: make([]Vec3, 0, len(tangents)),
		Faces:    make([]FaceT, 0, len(faces)),
	}

	m.Verts = append(m.Verts, verts...)
	for i, vert := range m.Verts {
		key := vert.toKey()
		m.uniqueVerts[key] = VertIndexT(i)
	}

	m.Normals = append(m.Normals, normals...)
	m.Tangents = append(m.Tangents, tangents...)
	for _, face := range faces {
		faceVerts := make([]Vec3, 0, len(face))
		for _, vertIdx := range face {
			faceVerts = append(faceVerts, verts[vertIdx])
		}
		m.AddFace(faceVerts)
	}

	return m
}

func meshClone(ls *lua.LState) int {
	orig := checkMesh(ls, 1)
	m := newMeshFrom(orig.Verts, orig.Normals, orig.Tangents, orig.Faces)

	ud := ls.NewUserData()
	ud.Value = m
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

// NewPolygonFromPoints creates a new mesh from points.
func NewPolygonFromPoints(pts []Vec3) *Mesh {
	m := newMeshFrom(pts, nil, nil, nil)
	m.AddFace(pts)
	return m
}

// NewMeshFromPolygons creates a new mesh from points.
func NewMeshFromPolygons(verts []Vec3, faces []FaceT) *Mesh {
	return newMeshFrom(verts, nil, nil, faces)
}

// NewMeshFromLineWithNormals creates a new mesh from points, normals, and tangents.
func NewMeshFromLineWithNormals(points, normals, tangents []Vec3) *Mesh {
	return newMeshFrom(points, normals, tangents, nil)
}

// NewMeshFromLine creates a new mesh from two points, divided into numSegs.
func NewMeshFromLine(v1, v2 *Vec3, numSegs int) *Mesh {
	// log.Printf("NewMeshFromLine: 2 points, %v segments", numSegs)
	verts := make([]Vec3, 0, numSegs+1)
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
		verts = append(verts, v)
	}
	verts = append(verts, *v2)
	return newMeshFrom(verts, nil, nil, nil)
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
		Verts:       make([]Vec3, 0, numVerts*len(backbone.Verts)),
		uniqueVerts: make(map[vertKeyT]VertIndexT, numVerts*len(backbone.Verts)),
		Faces:       make([]FaceT, 0, numVerts*(len(backbone.Verts)-1)),
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
			addedVertIdx := m.AddVert(v.Xform(xform))
			if addedVertIdx != VertIndexT(vIdx+i) {
				log.Fatalf("NewMeshFromExtrudeAlongCurve: programming error: addedVertIdx(%v) != vIdx(%v)+i(%v)", addedVertIdx, vIdx, i)
			}

			// log.Printf("verts[%v]=%v", len(m.Verts)-1, m.Verts[len(m.Verts)-1])
			if bvi == 0 {
				continue
			}

			// create a new quad for each extruded crossSection vertex
			if flip != 0 {
				m.Faces = append(m.Faces, FaceT{
					VertIndexT(vIdx + i),
					VertIndexT(vIdx + i - numVerts),
					VertIndexT(vIdx + ((i + 1) % numVerts) - numVerts),
					VertIndexT(vIdx + ((i + 1) % numVerts)),
				})
			} else {
				m.Faces = append(m.Faces, FaceT{
					VertIndexT(vIdx + i - numVerts),
					VertIndexT(vIdx + i),
					VertIndexT(vIdx + ((i + 1) % numVerts)),
					VertIndexT(vIdx + ((i + 1) % numVerts) - numVerts),
				})
			}
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
