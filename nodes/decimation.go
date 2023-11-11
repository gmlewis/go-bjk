package nodes

import (
	"fmt"
	"log"

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
	onFaceIdx      int
}

func (h *halfEdgeT) String() string {
	return fmt.Sprintf("{n: %v, toVertIdx: %v, onFaceIdx: %v}", h.edgeUnitNormal, h.toVertIdx, h.onFaceIdx)
}

// nextNonManifoldVerts finds the next collection of halfEdges that
// are non-manifold and returns them, or nil if none are found.
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

	return maps.Values(seenNormals[resultKey])
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
		return []*halfEdgeT{
			{
				edgeUnitNormal: fi.m.Verts[nextVertIdx].Sub(v).Normalized(),
				toVertIdx:      nextVertIdx,
				onFaceIdx:      faceIdx,
			},
			{
				edgeUnitNormal: fi.m.Verts[lastVertIdx].Sub(v).Normalized(),
				toVertIdx:      lastVertIdx,
				onFaceIdx:      faceIdx,
			},
		}
	}
	return nil
}

/*
decimatePhase1: vertIdx=2 {0.50 -0.50 4.50}, faceIdxes=[0 3 5 6 9 10]
2023/11/11 08:51:58 face[0]={{-0.50 -0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50} {-0.50 -0.50 4.50}}
2023/11/11 08:51:58 face[3]={{-0.50 -0.50 4.50} {0.50 -0.50 4.50} {0.50 1.50 4.50} {-0.50 1.50 4.50}}
2023/11/11 08:51:58 face[5]={{0.50 1.50 4.50} {0.50 -0.50 4.50} {0.50 -0.50 3.50} {0.50 1.50 3.50}}
2023/11/11 08:51:58 face[6]={{0.50 -0.50 3.50} {2.50 -0.50 3.50} {2.50 -0.50 4.50} {0.50 -0.50 4.50}}
2023/11/11 08:51:58 face[9]={{0.50 -0.50 4.50} {2.50 -0.50 4.50} {2.50 0.50 4.50} {0.50 0.50 4.50}}
2023/11/11 08:51:58 face[10]={{0.50 0.50 4.50} {0.50 0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50}}
2023/11/11 08:51:58 faceNormal[0]={0.00000 -1.00000 0.00000}
2023/11/11 08:51:58 faceNormal[3]={-0.00000 0.00000 1.00000}
2023/11/11 08:51:58 faceNormal[5]={1.00000 0.00000 -0.00000}
2023/11/11 08:51:58 faceNormal[6]={0.00000 -1.00000 0.00000}
2023/11/11 08:51:58 faceNormal[9]={-0.00000 0.00000 1.00000}
2023/11/11 08:51:58 faceNormal[10]={-1.00000 0.00000 0.00000}
*/

func (fi *faceInfoT) decimatePhase2(vertIdx int) {
}
