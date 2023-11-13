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

// faceStr2FaceIdxMapT maps a face "signature" (e.g. "0 1 2 3") to a face index.
type faceStr2FaceIdxMapT map[string]faceIndexT

// sharedVertsMapT represents a collection of shared vertices and maps them back to src ([0]) and dst([1]) face indices.
type sharedVertsMapT map[VertIndexT][2][]faceIndexT

// sharedEdgesMapT represents a collection of shared edges and maps them back to src ([0]) and dst([1]) face indices.
type sharedEdgesMapT map[edgeT][2][]faceIndexT

// sharedFacesMapT represents a collection of shared faces (keyed by face "signature") and maps them back to
// src ([0]) and dst([1]) face index.
type sharedFacesMapT map[string][2]faceIndexT

type faceInfoT struct {
	m   *Mesh
	src *infoSetT
	dst *infoSetT
}

type infoSetT struct {
	faces           []FaceT
	faceNormals     []Vec3
	vert2Faces      vert2FacesMapT
	edges2Faces     edge2FacesMapT
	faceStr2FaceIdx faceStr2FaceIdxMapT
	badEdges        edge2FacesMapT
	badFaces        face2EdgesMapT
}

func (fi *faceInfoT) swapSrcAndDst() {
	fi.src, fi.dst = fi.dst, fi.src
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
	src := m.genFaceInfoForSet(srcFaces)
	dst := m.genFaceInfoForSet(dstFaces)

	return &faceInfoT{
		m:   m,
		src: src,
		dst: dst,
	}
}

func (m *Mesh) genFaceInfoForSet(faces []FaceT) *infoSetT {
	infoSet := &infoSetT{
		faces:           faces,
		faceNormals:     make([]Vec3, 0, len(faces)),
		vert2Faces:      vert2FacesMapT{}, // key=vertIdx, value=[]faceIdx
		edges2Faces:     edge2FacesMapT{},
		faceStr2FaceIdx: faceStr2FaceIdxMapT{},
		badEdges:        edge2FacesMapT{},
		badFaces:        face2EdgesMapT{},
	}

	for i, face := range faces {
		faceIdx := faceIndexT(i)
		infoSet.faceNormals = append(infoSet.faceNormals, m.CalcFaceNormal(face))
		infoSet.faceStr2FaceIdx[face.String()] = faceIdx
		for i, vertIdx := range face {
			infoSet.vert2Faces[vertIdx] = append(infoSet.vert2Faces[vertIdx], faceIdx)
			nextVertIdx := face[(i+1)%len(face)]
			edge := makeEdge(vertIdx, nextVertIdx)
			infoSet.edges2Faces[edge] = append(infoSet.edges2Faces[edge], faceIdx)
		}
	}

	// Now find the bad edges and move them to the badEdges map.
	for edge, faceIdxes := range infoSet.edges2Faces {
		if len(faceIdxes) != 2 {
			infoSet.badEdges[edge] = faceIdxes
			for _, faceIdx := range faceIdxes {
				infoSet.badFaces[faceIdx] = append(infoSet.badFaces[faceIdx], edge)
			}
		}
	}
	for edge := range infoSet.badEdges {
		delete(infoSet.edges2Faces, edge)
	}

	return infoSet
}

func (fi *faceInfoT) findSharedVEFs() (sharedVertsMapT, sharedEdgesMapT, sharedFacesMapT) {
	// premature optimization:
	// if len(fi.dstFaces) < len(fi.srcFaces) {
	// 	fi.swapSrcAndDst()
	// }

	sharedVerts := sharedVertsMapT{}
	for vertIdx, dstFaces := range fi.dst.vert2Faces {
		if srcFaces, ok := fi.src.vert2Faces[vertIdx]; ok {
			sharedVerts[vertIdx] = [2][]faceIndexT{srcFaces, dstFaces}
		}
	}

	sharedEdges := sharedEdgesMapT{}
	for edge, dstFaces := range fi.dst.edges2Faces {
		if srcFaces, ok := fi.src.edges2Faces[edge]; ok {
			sharedEdges[edge] = [2][]faceIndexT{srcFaces, dstFaces}
		}
	}

	sharedFaces := sharedFacesMapT{}
	for faceStr, dstFaceIdx := range fi.dst.faceStr2FaceIdx {
		if srcFaceIdx, ok := fi.src.faceStr2FaceIdx[faceStr]; ok {
			sharedFaces[faceStr] = [2]faceIndexT{srcFaceIdx, dstFaceIdx}
		}
	}

	return sharedVerts, sharedEdges, sharedFaces
}

func (m *Mesh) faceArea(face FaceT) float64 {
	if len(face) == 4 {
		v1 := m.Verts[face[1]].Sub(m.Verts[face[0]]).Length()
		v2 := m.Verts[face[2]].Sub(m.Verts[face[1]]).Length()
		log.Printf("faceArea %+v: %v", face, v1*v2)
		return v1 * v2
	}
	log.Fatalf("faceArea: not implemented yet for %+v", face)
	return 0
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
