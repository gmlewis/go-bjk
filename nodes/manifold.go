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

// edgeT represents an edge and is a sorted array of two vertex indices.
type edgeT [2]VertIndexT

// edge2FacesMapT represents a mapping from an edge to one or more face indices.
type edge2FacesMapT map[edgeT][]faceIndexT

// vert2FacesMapT respresents a mapping from a vertex index to face indices.
type vert2FacesMapT map[VertIndexT][]faceIndexT

// face2EdgesMapT represents a mapping from a face index to edges.
type face2EdgesMapT map[faceIndexT][]edgeT

type faceInfoT struct {
	m *Mesh

	srcFaces         []FaceT
	srcFaceNormals   []Vec3
	srcFacesFromVert vert2FacesMapT
	srcEdges2Faces   edge2FacesMapT
	srcBadEdges      edge2FacesMapT
	srcBadFaces      face2EdgesMapT

	dstFaces         []FaceT
	dstFaceNormals   []Vec3
	dstFacesFromVert vert2FacesMapT
	dstEdges2Faces   edge2FacesMapT
	dstBadEdges      edge2FacesMapT
	dstBadFaces      face2EdgesMapT
}

func (fi *faceInfoT) swapSrcAndDst() {
	fi.srcFaces, fi.dstFaces = fi.dstFaces, fi.srcFaces
	fi.srcFaceNormals, fi.dstFaceNormals = fi.dstFaceNormals, fi.srcFaceNormals
	fi.srcFacesFromVert, fi.dstFacesFromVert = fi.dstFacesFromVert, fi.srcFacesFromVert
	fi.srcEdges2Faces, fi.dstEdges2Faces = fi.dstEdges2Faces, fi.srcEdges2Faces
	fi.srcBadEdges, fi.dstBadEdges = fi.dstBadEdges, fi.srcBadEdges
	fi.srcBadFaces, fi.dstBadFaces = fi.dstBadFaces, fi.srcBadFaces
}

func makeEdge(v1, v2 VertIndexT) edgeT {
	if v1 == v2 {
		log.Fatalf("programming error: makeEdge(%v,%v)", v1, v2)
	}
	if v1 < v2 {
		return edgeT{v1, v2}
	}
	return edgeT{v2, v1} // swap
}

// genFaceInfo calculates the face normals for every src and dst face
// and generates a map of good and bad edges (mapped to their respective faces).
func (m *Mesh) genFaceInfo(dstFaces, srcFaces []FaceT) *faceInfoT {
	sfn, sffv, se2f, sbe, sbf := m.genFaceInfoForSet(srcFaces)
	dfn, dffv, de2f, dbe, dbf := m.genFaceInfoForSet(dstFaces)

	return &faceInfoT{
		m: m,

		srcFaces:         srcFaces,
		srcFaceNormals:   sfn,
		srcFacesFromVert: sffv,
		srcEdges2Faces:   se2f,
		srcBadEdges:      sbe,
		srcBadFaces:      sbf,

		dstFaces:         dstFaces,
		dstFaceNormals:   dfn,
		dstFacesFromVert: dffv,
		dstEdges2Faces:   de2f,
		dstBadEdges:      dbe,
		dstBadFaces:      dbf,
	}
}

func (m *Mesh) genFaceInfoForSet(faces []FaceT) (faceNormals []Vec3, facesFromVert vert2FacesMapT, edges2Faces, badEdges edge2FacesMapT, badFaces face2EdgesMapT) {
	faceNormals = make([]Vec3, 0, len(faces))
	facesFromVert = vert2FacesMapT{} // key=vertIdx, value=[]faceIdx
	edges2Faces = edge2FacesMapT{}
	badEdges = edge2FacesMapT{}
	badFaces = face2EdgesMapT{}

	for i, face := range faces {
		faceIdx := faceIndexT(i)
		faceNormals = append(faceNormals, m.CalcFaceNormal(face))
		for i, vertIdx := range face {
			facesFromVert[vertIdx] = append(facesFromVert[vertIdx], faceIdx)
			nextVertIdx := face[(i+1)%len(face)]
			edge := makeEdge(vertIdx, nextVertIdx)
			edges2Faces[edge] = append(edges2Faces[edge], faceIdx)
		}
	}

	// Now find the bad edges and move them to the badEdges map.
	for edge, faceIdxes := range edges2Faces {
		if len(faceIdxes) != 2 {
			badEdges[edge] = faceIdxes
			for _, faceIdx := range faceIdxes {
				badFaces[faceIdx] = append(badFaces[faceIdx], edge)
			}
		}
	}
	for edge := range badEdges {
		delete(edges2Faces, edge)
	}

	return faceNormals,
		facesFromVert,
		edges2Faces,
		badEdges,
		badFaces
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
