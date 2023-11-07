package nodes

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

// MakeManifold attempts (using random, unoptimized heuristics) to
// generate a topologically closed (manifold) mesh from the given mesh.
func (m *Mesh) MakeManifold() error {
	if len(m.Faces) == 0 {
		return errors.New("no faces in mesh")
	}

	// First, calculate the face normals for every face and generate a mapping
	// from every vertex to a list of face indices that use that vertex.
	faceNormals := make([]Vec3, 0, len(m.Faces))
	facesFromVert := map[int][]int{} // key=vertIdx, value=[]faceIdx
	for faceIdx, face := range m.Faces {
		faceNormals = append(faceNormals, m.CalcFaceNormal(faceIdx))
		for _, vertIdx := range face {
			facesFromVert[vertIdx] = append(facesFromVert[vertIdx], faceIdx)
		}
	}

	// For every vertex, test the intersections of its faces and take action if needed.
	for vertIdx := range m.Verts {
		m.testManifoldFaces(vertIdx, facesFromVert[vertIdx])
	}

	return nil
}

func (m *Mesh) testManifoldFaces(vertIdx int, faceIdxes []int) {
	v := m.Verts[vertIdx]
	log.Printf("\n\ntestManifoldFaces: vertIdx=%v {%0.2f %0.2f %0.2f}, faceIdxes=%+v", vertIdx, v.X, v.Y, v.Z, faceIdxes)
	for _, faceIdx := range faceIdxes {
		log.Printf("face[%v]=%v", faceIdx, m.dumpFace(faceIdx))
	}
}

func (m *Mesh) dumpFace(faceIdx int) string {
	verts := make([]string, 0, len(m.Faces[faceIdx]))
	for _, vertIdx := range m.Faces[faceIdx] {
		v := m.Verts[vertIdx]
		verts = append(verts, fmt.Sprintf("{%0.2f %0.2f %0.2f}", v.X, v.Y, v.Z))
	}
	return fmt.Sprintf("{%v}", strings.Join(verts, " "))
}
