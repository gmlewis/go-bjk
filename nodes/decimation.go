package nodes

import (
	"fmt"
	"log"
	"sort"

	"golang.org/x/exp/maps"
)

// decimateFaces focuses on a single vertex, finds all faces attached to it,
// then decimates any faces whose topology is non-manifold.
//
// It does this in two phases:
//
//  1. decimate all faces where an edge unit vector points to two other faces
//     that do not share the same ending vertex.
//  2. decimate all faces where an edge unit vector lies directly on an unrelated face.
//
// Note that faceInfo is updated throughout the process to keep it self-consistent.
func (fi *faceInfoT) decimateFaces(vertIdx int) {
	fi.decimatePhase1(vertIdx)
	fi.decimatePhase2(vertIdx)
}

func (fi *faceInfoT) decimatePhase1(vertIdx int) {
	v := fi.m.Verts[vertIdx]
	faceIdxes := fi.facesFromVert[vertIdx]
	log.Printf("\n\ndecimatePhase1: vertIdx=%v {%0.2f %0.2f %0.2f}, faceIdxes=%+v", vertIdx, v.X, v.Y, v.Z, faceIdxes)
	for _, faceIdx := range faceIdxes {
		log.Printf("face[%v]=%v", faceIdx, fi.m.dumpFace(faceIdx))
	}
	for _, faceIdx := range faceIdxes {
		log.Printf("faceNormal[%v]=%v", faceIdx, fi.faceNormals[faceIdx])
	}

	// fi.m.Faces[5] = nil  // DEBUGGING ONLY!!!
	// fi.m.Faces[10] = nil // DEBUGGING ONLY!!!
	// fi.m.Faces[9] = nil  // DEBUGGING ONLY!!!

	// step 1 - find all edge unit vectors (and their ending vertex indices)
	// with respect to this vertex.
	for {
		nonManiVerts := fi.nextNonManifoldVerts(vertIdx)
		for i, nmv := range nonManiVerts {
			log.Printf("fromVertIdx=%v, nonManiVerts[%v] = %v, (toVert=%v)", vertIdx, i, nmv, fi.m.Verts[nmv.toVertIdx])
		}
		log.Fatalf("nonManiVerts: got %v", len(nonManiVerts))
	}

}

type halfEdgeT struct {
	edgeUnitNormal Vec3
	toVertIdx      int
	length         float64
	onFaceIdx      int
}

func (h *halfEdgeT) String() string {
	return fmt.Sprintf("{n: %v, toVertIdx: %v, length=%v, onFaceIdx: %v}", h.edgeUnitNormal, h.toVertIdx, h.length, h.onFaceIdx)
}

// nextNonManifoldVerts finds the next collection of halfEdges that
// are non-manifold and returns them (sorted by edge length descending),
// or nil if none are found.
func (fi *faceInfoT) nextNonManifoldVerts(vertIdx int) []*halfEdgeT {
	seenNormals := map[string]map[int]*halfEdgeT{}
	var resultKey string
	for _, faceIdx := range fi.facesFromVert[vertIdx] {
		for _, halfEdge := range fi.getHalfEdges(vertIdx, faceIdx) {
			key := halfEdge.edgeUnitNormal.String()
			if _, ok := seenNormals[key]; !ok {
				seenNormals[key] = map[int]*halfEdgeT{}
			}
			seenNormals[key][halfEdge.toVertIdx] = halfEdge
			if resultKey == "" && len(seenNormals[key]) > 1 {
				resultKey = key
			}
		}
	}

	result := maps.Values(seenNormals[resultKey])
	sort.Slice(result, func(i, j int) bool { return result[i].length > result[j].length })

	return result
}

// getHalfEdges returns both halfEdges for the given vertIdx and face.
func (fi *faceInfoT) getHalfEdges(vertIdx, faceIdx int) []*halfEdgeT {
	v := fi.m.Verts[vertIdx]
	face := fi.m.Faces[faceIdx]
	for i, vIdx := range face {
		if vIdx != vertIdx {
			continue
		}
		nextVertIdx := face[(i+1)%len(face)]
		lastVertIdx := face[(i-1+len(face))%len(face)]
		nextEdge := fi.m.Verts[nextVertIdx].Sub(v)
		lastEdge := fi.m.Verts[lastVertIdx].Sub(v)
		return []*halfEdgeT{
			{
				edgeUnitNormal: nextEdge.Normalized(),
				toVertIdx:      nextVertIdx,
				length:         nextEdge.Length(),
				onFaceIdx:      faceIdx,
			},
			{
				edgeUnitNormal: lastEdge.Normalized(),
				toVertIdx:      lastVertIdx,
				length:         lastEdge.Length(),
				onFaceIdx:      faceIdx,
			},
		}
	}
	return nil
}

/*
decimatePhase1: vertIdx=2 {0.50 -0.50 4.50}, faceIdxes=[0 3 5 6 9 10]
2023/11/11 11:47:32 face[0]={{-0.50 -0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50} {-0.50 -0.50 4.50}}
2023/11/11 11:47:32 face[3]={{-0.50 -0.50 4.50} {0.50 -0.50 4.50} {0.50 1.50 4.50} {-0.50 1.50 4.50}}
2023/11/11 11:47:32 face[5]={{0.50 1.50 4.50} {0.50 -0.50 4.50} {0.50 -0.50 3.50} {0.50 1.50 3.50}}
2023/11/11 11:47:32 face[6]={{0.50 -0.50 3.50} {2.50 -0.50 3.50} {2.50 -0.50 4.50} {0.50 -0.50 4.50}}
2023/11/11 11:47:32 face[9]={{0.50 -0.50 4.50} {2.50 -0.50 4.50} {2.50 0.50 4.50} {0.50 0.50 4.50}}
2023/11/11 11:47:32 face[10]={{0.50 0.50 4.50} {0.50 0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50}}
2023/11/11 11:47:32 faceNormal[0]={0.00000 -1.00000 0.00000}
2023/11/11 11:47:32 faceNormal[3]={-0.00000 0.00000 1.00000}
2023/11/11 11:47:32 faceNormal[5]={1.00000 0.00000 -0.00000}
2023/11/11 11:47:32 faceNormal[6]={0.00000 -1.00000 0.00000}
2023/11/11 11:47:32 faceNormal[9]={-0.00000 0.00000 1.00000}
2023/11/11 11:47:32 faceNormal[10]={-1.00000 0.00000 0.00000}
2023/11/11 12:11:28 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaceIdx: 5}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 12:11:28 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaceIdx: 10}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:47:32 nonManiVerts: got 2

with faces[5]=nil:
2023/11/11 11:56:55 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaceIdx: 3}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 11:56:55 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaceIdx: 10}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:56:55 nonManiVerts: got 2

with faces[5]=nil and faces[10]=nil:
2023/11/11 11:58:23 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaceIdx: 3}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 11:58:23 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaceIdx: 9}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:58:23 nonManiVerts: got 2

with faces[5]=nil and faces[10]=nil and faces[9]=nil:
NONE FOUND - MAKES SENSE!!!
*/

func (fi *faceInfoT) decimatePhase2(vertIdx int) {
}
