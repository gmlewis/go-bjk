package nodes

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// edgeT represents an edge and is a sorted array of two vertex indices.
type edgeT [2]VertIndexT

// edgeToFacesMapT represents a mapping from an edge to one or more face indices.
type edgeToFacesMapT map[edgeT][]faceIndexT

// vertToFacesMapT respresents a mapping from a vertex index to face indices.
type vertToFacesMapT map[VertIndexT][]faceIndexT

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
	vertToFaces     vertToFacesMapT
	edgeToFaces     edgeToFacesMapT
	faceStr2FaceIdx faceStr2FaceIdxMapT
	badEdges        edgeToFacesMapT
	badFaces        face2EdgesMapT

	facesTargetedForDeletion map[faceIndexT]bool
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
		vertToFaces:     vertToFacesMapT{}, // key=vertIdx, value=[]faceIdx
		edgeToFaces:     edgeToFacesMapT{},
		faceStr2FaceIdx: faceStr2FaceIdxMapT{},
		badEdges:        edgeToFacesMapT{},
		badFaces:        face2EdgesMapT{},

		facesTargetedForDeletion: map[faceIndexT]bool{},
	}

	for i, face := range faces {
		faceIdx := faceIndexT(i)
		infoSet.faceNormals = append(infoSet.faceNormals, fi.m.CalcFaceNormal(face))
		infoSet.faceStr2FaceIdx[face.toKey()] = faceIdx
		for i, vertIdx := range face {
			infoSet.vertToFaces[vertIdx] = append(infoSet.vertToFaces[vertIdx], faceIdx)
			nextVertIdx := face[(i+1)%len(face)]
			edge := makeEdge(vertIdx, nextVertIdx)
			infoSet.edgeToFaces[edge] = append(infoSet.edgeToFaces[edge], faceIdx)
		}
	}

	// Now find the bad edges and move them to the badEdges map.
	for edge, faceIdxes := range infoSet.edgeToFaces {
		if len(faceIdxes) != 2 {
			infoSet.badEdges[edge] = faceIdxes
			for _, faceIdx := range faceIdxes {
				infoSet.badFaces[faceIdx] = append(infoSet.badFaces[faceIdx], edge)
			}
		}
	}
	for edge := range infoSet.badEdges {
		delete(infoSet.edgeToFaces, edge)
	}

	return infoSet
}

func (fi *faceInfoT) findSharedVEFs() (sharedVertsMapT, sharedEdgesMapT, sharedFacesMapT) {
	// premature optimization:
	// if len(fi.dstFaces) < len(fi.srcFaces) {
	// 	fi.swapSrcAndDst()
	// }

	sharedVerts := sharedVertsMapT{}
	for vertIdx, dstFaces := range fi.dst.vertToFaces {
		if srcFaces, ok := fi.src.vertToFaces[vertIdx]; ok {
			sharedVerts[vertIdx] = [2][]faceIndexT{srcFaces, dstFaces}
		}
	}

	sharedEdges := sharedEdgesMapT{}
	for edge, dstFaces := range fi.dst.edgeToFaces {
		if srcFaces, ok := fi.src.edgeToFaces[edge]; ok {
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

type edgeVectorT struct {
	edge        edgeT
	fromVertIdx VertIndexT
	toVertIdx   VertIndexT
	toSubFrom   Vec3
	length      float64
}

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on edge)
// and therefore is completely independent of the winding order of the face!
// In addition to the edge vector, it also returns the VertIndexT of the other vertex.
func (is *infoSetT) connectedEdgeVectorFromVertOnFace(vertIdx VertIndexT, edge edgeT, faceIdx faceIndexT) edgeVectorT {
	notVertIdx := edge[0]
	if notVertIdx == vertIdx {
		notVertIdx = edge[1]
	}

	m := is.faceInfo.m
	face := is.faces[faceIdx]
	for i, pIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if pIdx == vertIdx && nextIdx != notVertIdx {
			// log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, nextIdx=%v, returning (%v).Sub(%v)",
			//   vertIdx, edge, faceIdx, i, pIdx, nextIdx, m.Verts[nextIdx], m.Verts[vertIdx])
			toSubFrom := m.Verts[nextIdx].Sub(m.Verts[vertIdx])
			return edgeVectorT{
				edge:        makeEdge(vertIdx, nextIdx),
				fromVertIdx: vertIdx,
				toVertIdx:   nextIdx,
				toSubFrom:   toSubFrom,
				length:      toSubFrom.Length(),
			}
		}
		if pIdx == vertIdx {
			lastVertIdx := face[(i-1+len(face))%len(face)]
			// log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, lastVertIdx=%v, returning (%v).Sub(%v)",
			//   vertIdx, edge, faceIdx, i, pIdx, lastVertIdx, m.Verts[lastVertIdx], m.Verts[vertIdx])
			toSubFrom := m.Verts[lastVertIdx].Sub(m.Verts[vertIdx])
			return edgeVectorT{
				edge:        makeEdge(vertIdx, lastVertIdx),
				fromVertIdx: vertIdx,
				toVertIdx:   lastVertIdx,
				toSubFrom:   toSubFrom,
				length:      toSubFrom.Length(),
			}
		}
	}

	log.Fatalf("connectedEdgeVectorFromVertOnFace: programming error for face %+v", face)
	return edgeVectorT{}
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

// edgeLength returns an edge's length.
func (is *infoSetT) edgeLength(edge edgeT) float64 {
	m := is.faceInfo.m
	return m.Verts[edge[0]].Sub(m.Verts[edge[1]]).Length()
}

// getFaceSideEdges returns a slice of edge vectors that are connected to (but not on) this face.
func (is *infoSetT) getFaceSideEdgeVectors(baseFaceIdx faceIndexT) []edgeVectorT {
	face := is.faces[baseFaceIdx]
	result := make([]edgeVectorT, 0, len(face))
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		edge := makeEdge(vertIdx, nextIdx)
		facesFromEdge := is.edgeToFaces[edge]
		for _, otherFaceIdx := range facesFromEdge {
			if otherFaceIdx == baseFaceIdx {
				continue
			}
			// take one edge from each connected face
			ev := is.connectedEdgeVectorFromVertOnFace(vertIdx, edge, otherFaceIdx)
			result = append(result, ev)
			break
		}
	}

	return result
}

// replaceFaceVertIdx finds and replaces the vertIdx on a face.
func (is *infoSetT) replaceFaceVertIdx(faceIdx faceIndexT, fromVertIdx, toVertIdx VertIndexT) {
	face := is.faces[faceIdx]
	for i, vertIdx := range face {
		if vertIdx == fromVertIdx {
			face[i] = toVertIdx
			return
		}
	}
}

func reverseMapFaceIndicesToEdges(sharedEdges sharedEdgesMapT) (srcFaceIndicesToEdges, dstFaceIndicesToEdges face2EdgesMapT) {
	srcFaceIndicesToEdges, dstFaceIndicesToEdges = face2EdgesMapT{}, face2EdgesMapT{}
	for edge, v := range sharedEdges {
		for _, faceIdx := range v[0] {
			srcFaceIndicesToEdges[faceIdx] = append(srcFaceIndicesToEdges[faceIdx], edge)
		}
		for _, faceIdx := range v[1] {
			dstFaceIndicesToEdges[faceIdx] = append(dstFaceIndicesToEdges[faceIdx], edge)
		}
	}
	return srcFaceIndicesToEdges, dstFaceIndicesToEdges
}

func reverseMapBadEdges(badEdges edgeToFacesMapT) (faceIndicesToEdges face2EdgesMapT) {
	faceIndicesToEdges = face2EdgesMapT{}
	for edge, faceIndices := range badEdges {
		for _, faceIdx := range faceIndices {
			faceIndicesToEdges[faceIdx] = append(faceIndicesToEdges[faceIdx], edge)
		}
	}
	return faceIndicesToEdges
}

// faceIndicesByEdgeCount returns a map of edge count to slice of faceIndices.
// So a face that has 6 shared edges would appear in the slice in result[6].
func faceIndicesByEdgeCount(inMap face2EdgesMapT) map[int][]faceIndexT {
	result := map[int][]faceIndexT{}
	for faceIdx, edges := range inMap {
		result[len(edges)] = append(result[len(edges)], faceIdx)
	}
	return result
}

// deleteFace deletes the face at the provided index, thereby shifting the other
// face indices around it! Always delete from last to first when deleting multiple faces.
// Do not call this directly. Let deleteFacesLastToFirst actually delete the faces.
func (is *infoSetT) deleteFace(deleteFaceIdx faceIndexT) {
	// log.Printf("\n\nDELETING FACE!!! %v", is.faceInfo.m.dumpFace(deleteFaceIdx, is.faces[deleteFaceIdx]))
	is.faces = slices.Delete(is.faces, int(deleteFaceIdx), int(deleteFaceIdx+1)) // invalidates other faceInfoT maps - last step.
}

// faceHasEdge checks that the given face has the provided edge.
func (is *infoSetT) faceHasEdge(faceIdx faceIndexT, edge edgeT) bool {
	face := is.faces[faceIdx]
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if edge[0] == vertIdx && edge[1] == nextIdx {
			return true
		}
		if edge[1] == vertIdx && edge[0] == nextIdx {
			return true
		}
	}
	return false
}

// deleteFacesLastToFirst deletes faces by sorting their indices, then deleting them highest to lowest.
func (is *infoSetT) deleteFacesLastToFirst(facesToDeleteMap map[faceIndexT]bool) {
	facesToDelete := maps.Keys(facesToDeleteMap)
	sort.Slice(facesToDelete, func(i, j int) bool { return facesToDelete[i] > facesToDelete[j] })
	for _, faceIdx := range facesToDelete {
		is.deleteFace(faceIdx)
	}
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
		// log.Printf("faceArea %+v: %v", face, v1*v2)
		return v1 * v2
	}
	log.Fatalf("faceArea: not implemented yet for face %+v with %v vertices", face, len(face))
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
