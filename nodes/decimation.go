package nodes

import (
	"fmt"
	"log"
	"math"
	"sort"

	"golang.org/x/exp/maps"
)

const (
	epsilon = 1e-5
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
		log.Printf("face[%v]=%+v=%v", faceIdx, fi.m.Faces[faceIdx], fi.m.dumpFace(faceIdx))
	}
	for _, faceIdx := range faceIdxes {
		log.Printf("faceNormal[%v]=%v", faceIdx, fi.faceNormals[faceIdx])
	}

	// fi.m.Faces[5] = nil  // DEBUGGING ONLY!!!
	// fi.m.Faces[10] = nil // DEBUGGING ONLY!!!
	// fi.m.Faces[9] = nil  // DEBUGGING ONLY!!!

	for {
		// step 1 - find all non-manifold halfEdges with respect to this vertex.
		nonManis := fi.nextNonManifoldHalfEdges(vertIdx)
		if len(nonManis) == 0 {
			return
		}

		// step 2 - decimate the longest faces common to this vertex along with
		// all connected topology to these faces.
		fi.decimateFacesBy(vertIdx, nonManis)

		log.Fatalf("nonManis: got %v", len(nonManis))
	}
}

type halfEdgeT struct {
	edgeUnitNormal Vec3
	toVertIdx      int
	length         float64
	onFaces        []int
}

func (h *halfEdgeT) String() string {
	return fmt.Sprintf("{edgeUnitNormal: %v, toVertIdx: %v, length=%v, onFaces: %+v}", h.edgeUnitNormal, h.toVertIdx, h.length, h.onFaces)
}

// nextNonManifoldHalfEdges finds the next collection of halfEdges that
// are non-manifold and returns them (sorted by edge length descending),
// or nil if none are found.
func (fi *faceInfoT) nextNonManifoldHalfEdges(vertIdx int) []*halfEdgeT {
	seenNormals := map[string]map[int]*halfEdgeT{}
	var resultKey string
	for _, faceIdx := range fi.facesFromVert[vertIdx] {
		for _, halfEdge := range fi.getHalfEdges(vertIdx, faceIdx) {
			key := halfEdge.edgeUnitNormal.String()
			if sn, ok := seenNormals[key]; ok {
				if oldHalfEdge, ok := sn[halfEdge.toVertIdx]; ok {
					oldHalfEdge.onFaces = append(oldHalfEdge.onFaces, faceIdx)
				} else {
					sn[halfEdge.toVertIdx] = halfEdge
				}
			} else {
				seenNormals[key] = map[int]*halfEdgeT{halfEdge.toVertIdx: halfEdge}
			}
			if resultKey == "" && len(seenNormals[key]) > 1 {
				resultKey = key
			}
		}
	}

	result := maps.Values(seenNormals[resultKey]) // will be nil if resultKey==""
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
				onFaces:        []int{faceIdx},
			},
			{
				edgeUnitNormal: lastEdge.Normalized(),
				toVertIdx:      lastVertIdx,
				length:         lastEdge.Length(),
				onFaces:        []int{faceIdx},
			},
		}
	}
	return nil
}

func (fi *faceInfoT) decimateFacesBy(vertIdx int, nonManis []*halfEdgeT) {
	for i, nmv := range nonManis {
		log.Printf("fromVertIdx=%v, nonManis[%v] = %v, (toVert=%v)", vertIdx, i, nmv, fi.m.Verts[nmv.toVertIdx])
	}
	if len(nonManis) > 2 {
		log.Fatalf("TODO: decimateFacesBy not yet implemented for %v nonManis", len(nonManis))
	}

	facesToCut := nonManis[0].onFaces
	cuttingVertIdx := nonManis[1].toVertIdx
	cuttingFaces := nonManis[1].onFaces
	for _, faceIdx := range facesToCut {
		if fi.cutFaceUsing(faceIdx, cuttingVertIdx, cuttingFaces) {
			continue
		}

		fi.splitOppositeFaceEdge(faceIdx, cuttingVertIdx, vertIdx, nonManis[0], nonManis[1])
	}
}

// cutFaceUsing attempts to cut a face by reusing existing vertices.
// It returns true on success.
func (fi *faceInfoT) cutFaceUsing(cutFaceIdx, cuttingVertIdx int, cuttingFaces []int) bool {
	log.Printf("cutFaceUsing: cutFaceIdx=%v, cuttingVertIdx=%v, cuttingFaces=%+v", cutFaceIdx, cuttingVertIdx, cuttingFaces)

	myVerts := map[int]bool{}
	for _, vertIdx := range fi.m.Faces[cutFaceIdx] {
		myVerts[vertIdx] = true
	}

	// step 1 - see if any of the cutting faces have any other verts that also lie on
	// (but are not shared by) this face  (in addition to the cuttingVertIdx).
	seenVertIdxes := map[int]bool{cuttingVertIdx: true}
	var vertsToCheck []int
	for _, cuttingFaceIdx := range cuttingFaces {
		cuttingFace := fi.m.Faces[cuttingFaceIdx]
		for _, vertIdx := range cuttingFace {
			if myVerts[vertIdx] || seenVertIdxes[vertIdx] {
				continue
			}
			seenVertIdxes[vertIdx] = true
			vertsToCheck = append(vertsToCheck, vertIdx)
		}
	}

	cuttingVertIdxes := fi.vertsLieOnFaceEdge(vertsToCheck, cutFaceIdx)
	if len(cuttingVertIdxes) == 0 {
		return false
	}

	for _, vertIdx := range cuttingVertIdxes {
		log.Printf("Found cutting vert: verts[%v]=%v", vertIdx, fi.m.Verts[vertIdx])
	}
	log.Fatalf("Found cutting verts: %+v", cuttingVertIdxes)
	// TODO
	return true
}

func (fi *faceInfoT) splitOppositeFaceEdge(cutFaceIdx, cuttingVertIdx, fromVertIdx int, edgeToCut, cuttingEdge *halfEdgeT) {
	log.Printf("splitOppositeFaceEdge: cutFaceIdx=%v, cuttingVertIdx=%v, fromVertIdx=%v\nedgeToCut=%v\ncuttingEdge=%v", cutFaceIdx, cuttingVertIdx, fromVertIdx, edgeToCut, cuttingEdge)
	// cutVector :=
}

func (fi *faceInfoT) vertsLieOnFaceEdge(vertsToCheck []int, faceIdx int) []int {
	var result []int

	face := fi.m.Faces[faceIdx]
	ignoreVerts := map[int]bool{}
	for _, vertIdx := range face {
		ignoreVerts[vertIdx] = true
	}

	for i, vertIdx := range face {
		p1 := fi.m.Verts[face[(i+1)%len(face)]].Sub(fi.m.Verts[vertIdx])
		pOnP1 := genPOnP1Func(p1)

		for _, pIdx := range vertsToCheck {
			if ignoreVerts[pIdx] {
				continue
			}

			p := fi.m.Verts[pIdx].Sub(fi.m.Verts[vertIdx])
			if pOnP1(p) {
				log.Printf("Found vert[%v]=%v on face[%v]=%+v!!!", pIdx, fi.m.Verts[pIdx], faceIdx, face)
				result = append(result, pIdx)
			}
		}
	}

	return result
}

func aboutEq(a, b float64) bool { return math.Abs(a-b) < epsilon }

func genPOnP1Func(p1 Vec3) func(p Vec3) bool {
	switch {
	case p1.X != 0 && p1.Y != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			ty := p.Y / p1.Y
			tz := p.Z / p1.Z
			return p.X > 0 && tx < 1 && p.Y > 0 && ty < 1 && p.Z > 0 && tz < 1 && aboutEq(tx, ty) && aboutEq(ty, tz)
		}
	case p1.X != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			tz := p.Z / p1.Z
			return p.X > 0 && tx < 1 && p.Z > 0 && tz < 1 && aboutEq(p.Y, 0) && aboutEq(tx, tz)
		}
	case p1.X != 0 && p1.Y != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			ty := p.Y / p1.Y
			return p.X > 0 && tx < 1 && p.Y > 0 && ty < 1 && aboutEq(p.Z, 0) && aboutEq(tx, ty)
		}
	case p1.Y != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			ty := p.Y / p1.Y
			tz := p.Z / p1.Z
			return p.Y > 0 && ty < 1 && p.Z > 0 && tz < 1 && aboutEq(p.X, 0) && aboutEq(ty, tz)
		}
	case p1.X != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			return p.X > 0 && tx < 1 && aboutEq(p.Y, 0) && aboutEq(p.Z, 0)
		}
	case p1.Y != 0:
		return func(p Vec3) bool {
			ty := p.Y / p1.Y
			return p.Y > 0 && ty < 1 && aboutEq(p.X, 0) && aboutEq(p.Z, 0)
		}
	case p1.Z != 0:
		return func(p Vec3) bool {
			tz := p.Z / p1.Z
			return p.Z > 0 && tz < 1 && aboutEq(p.X, 0) && aboutEq(p.Y, 0)
		}
	default:
		log.Fatalf("programming error: p1=%v", p1)
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
2023/11/11 12:31:22 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaces: [3 5]}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 12:31:22 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaces: [9 10]}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:47:32 nonManiVerts: got 2
2023/11/11 18:18:44 cutFaceUsing: cutFaceIdx=3, cuttingVertIdx=11, cuttingFaces=[9 10]
2023/11/11 18:18:44 cutFaceUsing: cutFaceIdx=5, cuttingVertIdx=11, cuttingFaces=[9 10]
2023/11/11 18:18:44 Found vert[10]={0.50000 0.50000 3.50000} on face[5]=[6 2 1 7]!!!
2023/11/11 18:18:44 Found cutting vert: verts[10]={0.50000 0.50000 3.50000}
2023/11/11 18:18:44 Found cutting verts: [10]

with faces[5]=nil:
2023/11/11 11:56:55 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaces: [3 5]}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 11:56:55 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaces: [9 10]}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:56:55 nonManiVerts: got 2

with faces[5]=nil and faces[10]=nil:
2023/11/11 11:58:23 fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaces: [3 5]}, (toVert={0.50000 1.50000 4.50000})
2023/11/11 11:58:23 fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaces: [9 10]}, (toVert={0.50000 0.50000 4.50000})
2023/11/11 11:58:23 nonManiVerts: got 2

with faces[5]=nil and faces[10]=nil and faces[9]=nil:
NONE FOUND - MAKES SENSE!!!
*/

func (fi *faceInfoT) decimatePhase2(vertIdx int) {
}
