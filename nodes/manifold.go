package nodes

import (
	"fmt"
	"log"
	"strings"
)

// func (m *Mesh) makeManifold() error {
// 	if len(m.Faces) == 0 {
// 		return errors.New("no faces in mesh")
// 	}
//
// 	faceInfo := m.genFaceInfo()
//
// 	for {
// 		faceInfo.changesMade = false
// 		for vertIdx := range m.Verts {
// 			faceInfo.decimateFaces(vertIdx)
// 		}
//
// 		if !faceInfo.changesMade {
// 			break
// 		}
// 	}
//
// 	return nil
// }

type faceInfoT struct {
	m *Mesh

	srcFaces         []FaceT
	srcFaceNormals   []Vec3
	srcFacesFromVert map[int][]int
	srcEdges2Faces   edge2FacesMapT
	srcBadEdges      edge2FacesMapT

	dstFaces         []FaceT
	dstFaceNormals   []Vec3
	dstFacesFromVert map[int][]int
	dstEdges2Faces   edge2FacesMapT
	dstBadEdges      edge2FacesMapT
}

func (fi *faceInfoT) swapSrcAndDst() {
	fi.srcFaces, fi.dstFaces = fi.dstFaces, fi.srcFaces
	fi.srcFaceNormals, fi.dstFaceNormals = fi.dstFaceNormals, fi.srcFaceNormals
	fi.srcFacesFromVert, fi.dstFacesFromVert = fi.dstFacesFromVert, fi.srcFacesFromVert
	fi.srcEdges2Faces, fi.dstEdges2Faces = fi.dstEdges2Faces, fi.srcEdges2Faces
	fi.srcBadEdges, fi.dstBadEdges = fi.dstBadEdges, fi.srcBadEdges
}

// edgeT represents an edge and is a sorted array of two vertex indices.
type edgeT [2]int

func makeEdge(v1, v2 int) edgeT {
	if v1 == v2 {
		log.Fatalf("programming error: makeEdge(%v,%v)", v1, v2)
	}
	if v1 < v2 {
		return [2]int{v1, v2}
	}
	return [2]int{v2, v1} // swap
}

// edge2FacesMapT represents a mapping from an edge to one or more face indices.
type edge2FacesMapT map[edgeT][]int

// genFaceInfo calculates the face normals for every src and dst face
// and generates a map of good and bad edges (mapped to their respective faces).
func (m *Mesh) genFaceInfo(dstFaces, srcFaces []FaceT) *faceInfoT {
	sfn, sffv, se2f, sbe := m.genFaceInfoForSet(srcFaces)
	dfn, dffv, de2f, dbe := m.genFaceInfoForSet(dstFaces)

	return &faceInfoT{
		m: m,

		srcFaces:         srcFaces,
		srcFaceNormals:   sfn,
		srcFacesFromVert: sffv,
		srcEdges2Faces:   se2f,
		srcBadEdges:      sbe,

		dstFaces:         dstFaces,
		dstFaceNormals:   dfn,
		dstFacesFromVert: dffv,
		dstEdges2Faces:   de2f,
		dstBadEdges:      dbe,
	}
}

func (m *Mesh) genFaceInfoForSet(faces []FaceT) (faceNormals []Vec3, facesFromVert map[int][]int, edges2Faces, badEdges edge2FacesMapT) {
	faceNormals = make([]Vec3, 0, len(faces))
	facesFromVert = map[int][]int{} // key=vertIdx, value=[]faceIdx
	edges2Faces = edge2FacesMapT{}
	badEdges = edge2FacesMapT{}

	for faceIdx, face := range faces {
		faceNormals = append(faceNormals, m.CalcFaceNormal(face))
		for i, vertIdx := range face {
			facesFromVert[vertIdx] = append(facesFromVert[vertIdx], faceIdx)
			nextVertIdx := face[(i+1)%len(face)]
			edge := makeEdge(vertIdx, nextVertIdx)
			edges2Faces[edge] = append(edges2Faces[edge], faceIdx)
		}
	}

	// Now find the bad edges and move them to the badEdges map.
	for edge, faces := range edges2Faces {
		if len(faces) != 2 {
			badEdges[edge] = faces
		}
	}
	for edge := range badEdges {
		delete(edges2Faces, edge)
	}

	return faceNormals,
		facesFromVert,
		edges2Faces,
		badEdges
}

func (m *Mesh) dumpFaces(faces []FaceT) string {
	var lines []string
	for i, face := range faces {
		lines = append(lines, fmt.Sprintf("face[%v]={%+v}: %v", i, face, m.dumpFace(face)))
	}
	return strings.Join(lines, "\n")
}

func (m *Mesh) dumpFace(face FaceT) string {
	verts := make([]string, 0, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx]
		verts = append(verts, fmt.Sprintf("{%0.2f %0.2f %0.2f}", v.X, v.Y, v.Z))
	}
	return fmt.Sprintf("{%v}", strings.Join(verts, " "))
}
