package nodes

import (
	"errors"
	"fmt"
	"strings"
)

// MakeManifold attempts (using random, unoptimized heuristics) to
// generate a topologically closed (manifold) mesh from the given mesh.
func (m *Mesh) MakeManifold() error {
	if len(m.Faces) == 0 {
		return errors.New("no faces in mesh")
	}

	faceInfo := m.genFaceInfo()

	for vertIdx := range m.Verts {
		// DEBUGGING ONLY!!!
		if vertIdx != 2 {
			continue
		}

		faceInfo.decimateFaces(vertIdx)
	}

	return nil
}

type faceInfoT struct {
	m             *Mesh
	allVertIdxes  map[string]int
	faceNormals   []Vec3
	facesFromVert map[int][]int
}

// genFaceInfo calculates the face normals for every face and generate a mapping
// from every vertex to a list of face indices that use that vertex.
func (m *Mesh) genFaceInfo() *faceInfoT {
	faceNormals := make([]Vec3, 0, len(m.Faces))
	facesFromVert := map[int][]int{} // key=vertIdx, value=[]faceIdx

	for faceIdx, face := range m.Faces {
		faceNormals = append(faceNormals, m.CalcFaceNormal(faceIdx))
		for _, vertIdx := range face {
			facesFromVert[vertIdx] = append(facesFromVert[vertIdx], faceIdx)
		}
	}

	allVertIdxes := make(map[string]int, len(m.Verts))
	for vertIdx, vert := range m.Verts {
		allVertIdxes[vert.String()] = vertIdx
	}

	return &faceInfoT{
		m:             m,
		allVertIdxes:  allVertIdxes,
		faceNormals:   faceNormals,
		facesFromVert: facesFromVert,
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
