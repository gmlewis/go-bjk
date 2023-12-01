package nodes

import (
	"fmt"
	"log"
	"sort"

	lua "github.com/yuin/gopher-lua"
)

const (
	vertSnappingResolution = "%0.3f %0.3f %0.3f"
)

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	// Do not manually add to Verts. Use AddVert instead.
	Verts       []Vec3
	uniqueVerts map[vertKeyT]VertIndexT

	Normals  []Vec3  // optional - per-vert normals
	Tangents []Vec3  // optional - per-vert tangents
	Faces    []FaceT // optional - when used, Normals and Tangents are unused.

	// these values are used by LerpAlongCurve to speed up future calculations
	segLengths  []float64
	totalLength float64
	tVals       []float64
}

// copyVertsFaces performs a deep copy of only the Verts and Faces.
func (m *Mesh) copyVertsFaces() (dup *Mesh) {
	dup = &Mesh{
		Verts:       append([]Vec3{}, m.Verts...), // Vec3 is a struct value - OK to copy.
		uniqueVerts: make(map[vertKeyT]VertIndexT, len(m.uniqueVerts)),
		Faces:       make([]FaceT, 0, len(m.Faces)), // FaceT is a slice - need to make a deep copy.
	}
	for _, face := range m.Faces {
		dup.Faces = append(dup.Faces, append(FaceT{}, face...))
	}
	for k, v := range m.uniqueVerts {
		dup.uniqueVerts[k] = v
	}
	return dup
}

// VertIndexT represents a vertex index.
type VertIndexT int

// FaceT represents a face and is a slice of vertex indices.
type FaceT []VertIndexT

// vertKeyT represents a vertex key (or "signature") which uniquely identifies a vertex for easy merging.
type vertKeyT string

// toKey generates a vertex vertKeyT (or "signature") which is a string representation of the vertex.
// Note that "positive zero" and "negative zero" map to different strings, so convert negative zeros to positive zeros.
// This essentially "snaps" vertices together that are within the "vertSnappingResolution".
// Note that since these keys are used in maps, they hash better without the surrounding curly braces {} or brackets [].
func (v Vec3) toKey() vertKeyT {
	if AboutEq(v.X, 0) {
		v.X = 0
	}
	if AboutEq(v.Y, 0) {
		v.Y = 0
	}
	if AboutEq(v.Z, 0) {
		v.Z = 0
	}
	return vertKeyT(fmt.Sprintf(vertSnappingResolution, v.X, v.Y, v.Z))
}

// faceKeyT represents a face key (or "signature") which uniquely identifies a face consisting of the same verts.
type faceKeyT string

// toKey generates a faceKeyT (or "signature") which is a string of the sorted vertex indices.
func (f FaceT) toKey() faceKeyT {
	verts := append([]VertIndexT{}, f...)
	sort.Slice(verts, func(i, j int) bool { return verts[i] < verts[j] })
	return faceKeyT(fmt.Sprintf("%v", verts))
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
// Note that some lua code could create a face where an
// edge has two vert indices are identical. AddFace
// prevents that from happening.
func (m *Mesh) AddFace(verts []Vec3) FaceT {
	face := make([]VertIndexT, 0, len(verts))
	for i, vert := range verts {
		vertIdx := m.AddVert(vert)
		if i > 0 && vertIdx == face[len(face)-1] {
			continue // prevent two identical consecutive vert indices.
		}
		face = append(face, vertIdx)
	}
	// now check that the ending vertex does not match the initial one.
	for len(face) > 1 {
		if face[0] != face[len(face)-1] {
			break
		}
		face = face[:len(face)-1]
	}
	if len(face) < 3 {
		log.Fatalf("programming error: AddFace: %+v, face=%v", verts, face)
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

// NewMesh returns a new, empty mesh.
func NewMesh() *Mesh {
	return newMeshFrom(nil, nil, nil, nil)
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

// NewLineFromPoints creates a new mesh with only verts.
func NewLineFromPoints(pts []Vec3) *Mesh {
	m := newMeshFrom(pts, nil, nil, nil)
	return m
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

// NewLine creates a new mesh from two points, divided into numSegs.
func NewLine(v1, v2 *Vec3, numSegs int) *Mesh {
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
func (m *Mesh) CalcFaceNormal(face FaceT) Vec3 {
	if len(m.Verts) < 3 || len(face) < 3 {
		log.Fatalf("CalcFaceNormal: want >=3 verts >=1 face, got %v total verts and %v verts in face (ignore face index):\n%v", len(m.Verts), len(face), m.dumpFace(-1, face))
	}

	var sum Vec3
	for i, vertIdx := range face {
		lastIdx := face[(i-1+len(face))%len(face)]
		nextIdx := face[(i+1)%len(face)]
		p1 := Vec3Sub(m.Verts[nextIdx], m.Verts[vertIdx])
		p0 := Vec3Sub(m.Verts[lastIdx], m.Verts[vertIdx])
		cross := Vec3Cross(p1, p0)
		sum = Vec3Add(sum, cross)
	}

	return sum.Normalized()
}

// LerpAlongCurve returns a Vec3 representing the percentage t (0 to 1) along a
// curve (the points in the mesh). It caches the length of the curve segments
// upon first use for later speedup.
func (m *Mesh) LerpAlongCurve(t float64) *Vec3 {
	if len(m.Verts) == 0 {
		return &Vec3{}
	}

	if len(m.segLengths) == 0 {
		m.segLengths = make([]float64, 0, len(m.Verts)-1)
		for i, vert := range m.Verts[:len(m.Verts)-1] {
			length := m.Verts[i+1].Sub(vert).Length()
			m.totalLength += length
			m.segLengths = append(m.segLengths, length)
		}
		m.tVals = make([]float64, 0, len(m.Verts))
		var length float64
		for i := range m.Verts[:len(m.Verts)] {
			if i == 0 {
				m.tVals = append(m.tVals, 0)
				continue
			}
			length += m.segLengths[i-1]
			m.tVals = append(m.tVals, length/m.totalLength)
		}
	}

	if t <= 0 {
		return &m.Verts[0]
	}
	if t >= 1 {
		return &m.Verts[len(m.Verts)-1]
	}

	for i, tVal := range m.tVals {
		if AboutEq(tVal, t) {
			return &m.Verts[i]
		}
		if tVal < t {
			continue
		}
		lastT := m.tVals[i-1]
		diff := tVal - lastT
		if diff == 0 {
			continue
		}
		frac := (t - lastT) / diff
		lastV := m.Verts[i-1]
		thisV := m.Verts[i]
		v := lastV.Add(thisV.Sub(lastV).MulScalar(frac))
		return &v
	}

	return &m.Verts[len(m.Verts)-1]
}
