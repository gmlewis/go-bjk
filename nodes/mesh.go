package nodes

import (
	"log"

	"github.com/gmlewis/go3d/float64/vec3"
	lua "github.com/yuin/gopher-lua"
)

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	Verts    []Vec3
	Normals  []Vec3  // optional - per-vert normals
	Tangents []Vec3  // optional - per-vert tangents
	Faces    [][]int // optional
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

func meshClone(ls *lua.LState) int {
	orig := checkMesh(ls, 1)
	m := &Mesh{
		Verts:    make([]Vec3, 0, len(orig.Verts)),
		Normals:  make([]Vec3, 0, len(orig.Normals)),
		Tangents: make([]Vec3, 0, len(orig.Tangents)),
		Faces:    make([][]int, 0, len(orig.Faces)),
	}

	for _, v := range orig.Verts {
		m.Verts = append(m.Verts, v)
	}
	for _, v := range orig.Normals {
		m.Normals = append(m.Normals, v)
	}
	for _, v := range orig.Tangents {
		m.Tangents = append(m.Tangents, v)
	}
	for _, v := range orig.Faces {
		m.Faces = append(m.Faces, append([]int{}, v...))
	}

	ud := ls.NewUserData()
	ud.Value = m
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaMeshTypeName))
	ls.Push(ud)
	return 1
}

// Merge merges src into dst for Ops.merge(dst, src).
func (dst *Mesh) Merge(src *Mesh) {
	// Currently, a naive merge is performed by not checking if any Verts are shared.
	verts := make([]Vec3, 0, len(dst.Verts)+len(src.Verts))
	normals := make([]Vec3, 0, len(dst.Normals)+len(src.Normals))
	tangents := make([]Vec3, 0, len(dst.Tangents)+len(src.Tangents))
	faces := make([][]int, 0, len(dst.Faces)+len(src.Faces))

	for _, v := range dst.Verts {
		verts = append(verts, v)
	}
	for _, v := range src.Verts {
		verts = append(verts, v)
	}

	for _, v := range dst.Normals {
		normals = append(normals, v)
	}
	for _, v := range src.Normals {
		normals = append(normals, v)
	}

	for _, v := range dst.Tangents {
		tangents = append(tangents, v)
	}
	for _, v := range src.Tangents {
		tangents = append(tangents, v)
	}

	for _, v := range dst.Faces {
		faces = append(faces, append([]int{}, v...))
	}
	numVerts := len(dst.Verts)
	adjFace := func(src []int) []int {
		result := make([]int, 0, len(src))
		for _, f := range src {
			result = append(result, f+numVerts)
		}
		return result
	}
	for _, v := range src.Faces {
		faces = append(faces, adjFace(v))
	}

	dst.Verts = verts
	dst.Normals = normals
	dst.Tangents = tangents
	dst.Faces = faces
}

// NewPolygonFromPoints creates a new mesh from points.
func NewPolygonFromPoints(pts []Vec3) *Mesh {
	m := &Mesh{Verts: pts, Faces: [][]int{make([]int, 0, len(pts))}}
	for i := 0; i < len(pts); i++ {
		m.Faces[0] = append(m.Faces[0], i)
	}
	return m
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

// NewMeshFromLine creates a new mesh from two points, divided into numSegs.
func NewMeshFromLine(v1, v2 *Vec3, numSegs int) *Mesh {
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
		log.Printf("GML: NewMeshFromExtrudeAlongCurve - calling generateTangents() - #Tangents=%v, #Verts=%v", len(backbone.Tangents), len(backbone.Verts))
		backbone.generateTangents()
	}

	log.Printf("GML: nmfeac: backbone.Tangents[0]=%v", backbone.Tangents[0])
	baseRotX, baseRotY, baseRotZ := backbone.Tangents[0].GetRotXYZ()
	// log.Printf("nmfeac: startPos=%v, baseRotX=%v°, baseRotY=%v°, baseRotZ=%v°",
	//     startPos, baseRotX*math.Pi/180, baseRotY*math.Pi/180, baseRotZ*math.Pi/180)

	// First, make a copy of the crossSection verts positioned in-place at the start of the backbone.
	for _, v := range crossSection.Verts {
		m.Verts = append(m.Verts, v.Add(startPos))
		// log.Printf("verts[%v]=%v", len(m.Verts)-1, m.Verts[len(m.Verts)-1])
	}

	// For each segment, add numVerts to the mesh, rotated and translated into place, and create new faces
	// that connect to the last set of numVerts.
	for bvi := 1; bvi < len(backbone.Verts); bvi++ {
		rotX, rotY, rotZ := backbone.Tangents[bvi].GetRotXYZ()
		rx, ry, rz := rotX-baseRotX, rotY-baseRotY, rotZ-baseRotZ

		bvert := backbone.Verts[bvi]
		xform := GenXform(rx, ry, rz, bvert)
		vIdx := len(m.Verts)
		// log.Printf("bvi=%v, bvert=%v, xform=%v, vIdx=%v", bvi, bvert, xform, vIdx)
		for i, v := range crossSection.Verts {
			m.Verts = append(m.Verts, v.Xform(xform))
			// log.Printf("verts[%v]=%v", len(m.Verts)-1, m.Verts[len(m.Verts)-1])
			// create a new quad for each extruded crossSection vertex
			m.Faces = append(m.Faces, []int{
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

func (m *Mesh) CalcNormal(faceIndex int) Vec3 {
	if len(m.Verts) < 3 || len(m.Faces) <= faceIndex || len(m.Faces[faceIndex]) < 3 {
		log.Fatalf("CalcNormal want >=3 points >=1 face, got %#v", *m)
	}
	face := m.Faces[faceIndex]
	v1, v2, v3 := m.Verts[face[0]], m.Verts[face[1]], m.Verts[face[2]]
	pt1, pt2, pt3 := vec3.T{v1.X, v1.Y, v1.Z}, vec3.T{v2.X, v2.Y, v2.Z}, vec3.T{v3.X, v3.Y, v3.Z}
	vec1 := vec3.Sub(&pt2, &pt1)
	vec2 := vec3.Sub(&pt3, &pt1)
	cross := vec3.Cross(&vec1, &vec2)
	return Vec3{X: cross[0], Y: cross[1], Z: cross[2]}
}
