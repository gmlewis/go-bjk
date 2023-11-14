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
type faceStr2FaceIdxMapT map[faceKeyT]faceIndexT

// sharedVertsMapT represents a collection of shared vertices and maps them back to src ([0]) and dst([1]) face indices.
type sharedVertsMapT map[VertIndexT][2][]faceIndexT

// sharedEdgesMapT represents a collection of shared edges and maps them back to src ([0]) and dst([1]) face indices.
type sharedEdgesMapT map[edgeT][2][]faceIndexT

// sharedFacesMapT represents a collection of shared faces (keyed by face "signature") and maps them back to
// src ([0]) and dst([1]) face index.
type sharedFacesMapT map[faceKeyT][2]faceIndexT

type faceInfoT struct {
	m   *Mesh
	src *infoSetT
	dst *infoSetT
}

type infoSetT struct {
	faceInfo        *faceInfoT
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
	fi := &faceInfoT{m: m}
	fi.src = fi.genFaceInfoForSet(srcFaces)
	fi.dst = fi.genFaceInfoForSet(dstFaces)
	return fi
}

func (fi *faceInfoT) genFaceInfoForSet(faces []FaceT) *infoSetT {
	infoSet := &infoSetT{
		faceInfo:        fi,
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
		infoSet.faceNormals = append(infoSet.faceNormals, fi.m.CalcFaceNormal(face))
		infoSet.faceStr2FaceIdx[face.toKey()] = faceIdx
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

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on edge)
// and therefore is completely independent of the winding order of the face!
// In addition to the edge vector, it also returns the VertIndexT of the other vertex.
func (is *infoSetT) connectedEdgeVectorFromVertOnFace(vertIdx VertIndexT, edge edgeT, faceIdx faceIndexT) (VertIndexT, Vec3) {
	notVertIdx := edge[0]
	if notVertIdx == vertIdx {
		notVertIdx = edge[1]
	}

	m := is.faceInfo.m
	face := is.faces[faceIdx]
	for i, pIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if pIdx == vertIdx && nextIdx != notVertIdx {
			log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, nextIdx=%v, returning (%v).Sub(%v)",
				vertIdx, edge, faceIdx, i, pIdx, nextIdx, m.Verts[nextIdx], m.Verts[vertIdx])
			return nextIdx, m.Verts[nextIdx].Sub(m.Verts[vertIdx])
		}
		if pIdx == vertIdx {
			lastVertIdx := face[(i-1+len(face))%len(face)]
			log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, lastVertIdx=%v, returning (%v).Sub(%v)",
				vertIdx, edge, faceIdx, i, pIdx, lastVertIdx, m.Verts[lastVertIdx], m.Verts[vertIdx])
			return lastVertIdx, m.Verts[lastVertIdx].Sub(m.Verts[vertIdx])
		}
	}

	log.Fatalf("connectedEdgeVectorFromVertOnFace: programming error for face %+v", face)
	return 0, Vec3{}
}

// This preserves the order of vertex indicies as they appear in the face definition.
func (is *infoSetT) getEdgeVertsInWindingOrder(edge edgeT, faceIdx faceIndexT) [2]VertIndexT {
	face := is.faces[faceIdx]
	for i, pIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if edge[0] == pIdx && edge[1] == nextIdx {
			return [2]VertIndexT{edge[0], edge[1]}
		}
		if edge[1] == pIdx && edge[0] == nextIdx {
			return [2]VertIndexT{edge[1], edge[0]}
		}
	}

	log.Fatalf("getEdgeVertsInWindingOrder: programming error: invalid edge %v for face %+v", edge, face)
	return [2]VertIndexT{}
}

// // edgeVector returns the vector representing this edge.
// // Note that the edge order does _NOT_ represent the winding order!
// // Therefore the original winding order needs to be found and preserved.
// func (is *infoSetT) edgeVector(edge edgeT, faceIdx faceIndexT) Vec3 {
// 	vertIdxes := is.getEdgeVertsInWindingOrder(edge, faceIdx)
// 	m := is.faceInfo.m
// 	return m.Verts[vertIdxes[1]].Sub(m.Verts[vertIdxes[0]])
// }

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
		lines = append(lines, m.dumpFace(faceIndexT(i), face))
	}
	return strings.Join(lines, "\n")
}

func (m *Mesh) dumpFace(faceIdx faceIndexT, face FaceT) string {
	verts := make([]string, 0, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx]
		verts = append(verts, v.String())
	}
	return fmt.Sprintf("face[%v]={%+v}: {%v}", faceIdx, face, strings.Join(verts, " "))
}
