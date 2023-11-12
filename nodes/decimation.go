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
	edgeUnitVector Vec3
	toVertIdx      int
	length         float64
	onFaces        []int
}

func (h *halfEdgeT) String() string {
	return fmt.Sprintf("{edgeUnitVector: %v, toVertIdx: %v, length=%v, onFaces: %+v}", h.edgeUnitVector, h.toVertIdx, h.length, h.onFaces)
}

// nextNonManifoldHalfEdges finds the next collection of halfEdges that
// are non-manifold and returns them (sorted by edge length descending),
// or nil if none are found.
func (fi *faceInfoT) nextNonManifoldHalfEdges(vertIdx int) []*halfEdgeT {
	seenNormals := map[string]map[int]*halfEdgeT{}
	var resultKey string
	for _, faceIdx := range fi.facesFromVert[vertIdx] {
		for _, halfEdge := range fi.getHalfEdges(vertIdx, faceIdx) {
			key := halfEdge.edgeUnitVector.String()
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
				edgeUnitVector: nextEdge.Normalized(),
				toVertIdx:      nextVertIdx,
				length:         nextEdge.Length(),
				onFaces:        []int{faceIdx},
			},
			{
				edgeUnitVector: lastEdge.Normalized(),
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

		fi.splitOppositeFaceEdge(faceIdx, cuttingVertIdx, vertIdx, nonManis[0].toVertIdx)
	}
}

func (fi *faceInfoT) splitOppositeFaceEdge(cutFaceIdx, cuttingVertIdx, fromVertIdx, toVertIdx int) {
	log.Printf("splitOppositeFaceEdge: cutFaceIdx=%v, cuttingVertIdx=%v, fromVertIdx=%v, toVertIdx=%v", cutFaceIdx, cuttingVertIdx, fromVertIdx, toVertIdx)
	cutVector := Vec3Cross(fi.faceNormals[cutFaceIdx], fi.m.Verts[toVertIdx].Sub(fi.m.Verts[fromVertIdx])).Normalized()
	log.Printf("cutVector from vert[%v]=%v: %v", cuttingVertIdx, fi.m.Verts[cuttingVertIdx], cutVector)

	cutInfo := fi.findOppositeEdge(cutFaceIdx, cuttingVertIdx, fromVertIdx, cutVector)
	if cutInfo == nil {
		log.Fatalf("splitOppositeFaceEdge: unable to find opposite face edge")
	}

	log.Printf("splitting face: cutInfo=%v", cutInfo)

	// inject new vertex if it doesn't already exist
	newVertKey := cutInfo.newCutVert.String()
	newVertIdx, ok := fi.allVertIdxes[newVertKey]
	if !ok {
		newVertIdx = len(fi.m.Verts)
		fi.allVertIdxes[newVertKey] = newVertIdx
		fi.m.Verts = append(fi.m.Verts, cutInfo.newCutVert)
	}

	oldFace := fi.m.Faces[cutFaceIdx]
	newFace1, newFace2 := fi.newFacesFromOld(oldFace, cuttingVertIdx, fromVertIdx, newVertIdx, cutInfo.fromVertIdx)
	log.Printf("\n\nsplitOppositeFaceEdge: newFace1=%v, newFace2=%v", newFace1, newFace2)

	fi.m.Faces[cutFaceIdx] = newFace1
	newFaceIdx := len(fi.m.Faces)
	fi.m.Faces = append(fi.m.Faces, newFace2)
	fi.faceNormals = append(fi.faceNormals, fi.faceNormals[cutFaceIdx]) // copy identical face normal
	fi.facesFromVert[newVertIdx] = []int{cutFaceIdx, newFaceIdx}
}

func (fi *faceInfoT) newFacesFromOld(oldFace []int, cuttingVertIdx, fromVertIdx, newVertIdx, toCutFromVertIdx int) (f1, f2 []int) {
	for i, vertIdx := range oldFace {
		f1 = append(f1, vertIdx)

		if vertIdx == fromVertIdx {
			f1 = append(f1, cuttingVertIdx)
			f1 = append(f1, newVertIdx)
			f2 = append(f2, cuttingVertIdx)
			addToF2 := true
			for j := 1; i+j < len(oldFace); j++ {
				vertIdx = oldFace[i+j]
				if addToF2 {
					f2 = append(f2, vertIdx)
				} else {
					f1 = append(f1, vertIdx)
				}
				if vertIdx == toCutFromVertIdx {
					f2 = append(f2, newVertIdx)
					addToF2 = false
				}
			}
			return f1, f2
		}
		if vertIdx == toCutFromVertIdx {
			f1 = append(f1, newVertIdx)
			f1 = append(f1, cuttingVertIdx)
			f2 = append(f2, newVertIdx)
			addToF2 := true
			for j := 1; i+j < len(oldFace); j++ {
				vertIdx = oldFace[i+j]
				if addToF2 {
					f2 = append(f2, vertIdx)
				} else {
					f1 = append(f1, vertIdx)
				}
				if vertIdx == fromVertIdx {
					f2 = append(f2, cuttingVertIdx)
					addToF2 = false
				}
			}
			return f1, f2
		}
	}

	log.Fatalf("programming error: newFacesFromOld(oldFace=%+v), f1=%+v, f2=%+v", oldFace, f1, f2)
	return nil, nil
}

type cutInfoT struct {
	fromVertIdx int
	newCutVert  Vec3
	toVertIdx   int
}

func (fi *faceInfoT) findOppositeEdge(cutFaceIdx, cuttingVertIdx, fromVertIdx int, cutVector Vec3) *cutInfoT {
	cuttingVert := fi.m.Verts[cuttingVertIdx]

	face := fi.m.Faces[cutFaceIdx]
	for i, vertIdx := range face {
		if vertIdx == fromVertIdx {
			continue
		}

		// https://web.archive.org/web/20180927042445/http://mathforum.org/library/drmath/view/62814.html
		p1 := fi.m.Verts[vertIdx]
		nextVertIdx := face[(i+1)%len(face)]
		edgeVector := fi.m.Verts[nextVertIdx].Sub(p1)
		lhs := Vec3Cross(edgeVector, cutVector)
		if lhs.AboutZero() {
			continue
		}
		rhs := Vec3Cross(cuttingVert.Sub(p1), cutVector)

		var ratio float64
		switch {
		case AboutEq(lhs.X, 0) && AboutEq(lhs.Y, 0):
			ratio = rhs.Z / lhs.Z
		case AboutEq(lhs.X, 0) && AboutEq(lhs.Z, 0):
			ratio = rhs.Y / lhs.Y
		case AboutEq(lhs.Y, 0) && AboutEq(lhs.Z, 0):
			ratio = rhs.X / lhs.X
		default:
			log.Fatalf("findOppositeEdge: unhandled case: lhs=%v, rhs=%v", lhs, rhs)
		}

		log.Printf("found opposite edge: i=%v, p1=%v=%v: lhs=%v, rhs=%v, ratio=%v", i, vertIdx, p1, lhs, rhs, ratio)
		return &cutInfoT{
			fromVertIdx: vertIdx,
			newCutVert:  p1.Add(edgeVector.MulScalar(ratio)),
			toVertIdx:   nextVertIdx,
		}
	}

	return nil
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
	// Also determine the 'fromVertIdx' of the cuttingVertIdx.
	seenVertIdxes := map[int]bool{cuttingVertIdx: true}
	vertsToCheck := []int{cuttingVertIdx}
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

	vertsOnEdgesInfo := fi.vertsLieOnFaceEdge(vertsToCheck, cutFaceIdx)
	if len(vertsOnEdgesInfo) < 1 {
		log.Fatalf("programming error: expected 1 result from vertsLieOnFaceEdge, got none")
	}

	if len(vertsOnEdgesInfo) == 1 {
		return false
	}

	if len(vertsOnEdgesInfo) > 2 {
		log.Fatalf("cutFaceUsing: unhandled case: found cutting verts: %+v", vertsOnEdgesInfo)
	}

	log.Printf("cutFaceUsing: vertsOnEdgesInfo[0]=%v, vertsOnEdgesInfo[1]=%v", vertsOnEdgesInfo[0], vertsOnEdgesInfo[1])

	fromVertIdx := vertsOnEdgesInfo[0].fromVertIdx
	if vertsOnEdgesInfo[0].vertOnEdgeIdx != cuttingVertIdx {
		log.Fatalf("cutFaceUsing: programming error: vertsOnEdgesInfo[0]=%v", vertsOnEdgesInfo[0])
	}

	newVertIdx := vertsOnEdgesInfo[1].vertOnEdgeIdx
	toCutFromVertIdx := vertsOnEdgesInfo[1].fromVertIdx
	log.Printf("cutFaceUsing: Found cutting vert: verts[%v]=%v", newVertIdx, fi.m.Verts[newVertIdx])

	oldFace := fi.m.Faces[cutFaceIdx]
	newFace1, newFace2 := fi.newFacesFromOld(oldFace, cuttingVertIdx, fromVertIdx, newVertIdx, toCutFromVertIdx)
	log.Printf("\n\ncutFaceUsing: newFace1=%v, newFace2=%v", newFace1, newFace2)

	fi.m.Faces[cutFaceIdx] = newFace1
	newFaceIdx := len(fi.m.Faces)
	fi.m.Faces = append(fi.m.Faces, newFace2)
	fi.faceNormals = append(fi.faceNormals, fi.faceNormals[cutFaceIdx]) // copy identical face normal
	fi.facesFromVert[newVertIdx] = []int{cutFaceIdx, newFaceIdx}

	return true
}

type vertOnEdgeInfoT struct {
	vertOnEdgeIdx int
	fromVertIdx   int
}

func (fi *faceInfoT) vertsLieOnFaceEdge(vertsToCheck []int, faceIdx int) []*vertOnEdgeInfoT {
	var result []*vertOnEdgeInfoT

	face := fi.m.Faces[faceIdx]
	ignoreVerts := map[int]bool{}
	for _, vertIdx := range face {
		ignoreVerts[vertIdx] = true
	}

	for i, vertIdx := range face {
		nextVertIdx := face[(i+1)%len(face)]
		// log.Printf("vertsLieOnFaceEdge: i=%v, looking at: v[%v]=%v to v[%v]=%v",
		//   i, vertIdx, fi.m.Verts[vertIdx], nextVertIdx, fi.m.Verts[nextVertIdx])
		p1 := fi.m.Verts[nextVertIdx].Sub(fi.m.Verts[vertIdx])
		pOnP1 := genPOnP1Func(p1)

		for _, pIdx := range vertsToCheck {
			if ignoreVerts[pIdx] {
				log.Printf("vertsLieOnFaceEdge: ignoring vert[%v]=%v", pIdx, fi.m.Verts[pIdx])
				continue
			}
			// log.Printf("vertsLieOnFaceEdge: i=%v, looking at: v[%v]=%v on line segment?",
			// 	i, pIdx, fi.m.Verts[pIdx])

			p := fi.m.Verts[pIdx].Sub(fi.m.Verts[vertIdx])
			if pOnP1(p) {
				log.Printf("vertsLieOnFaceEdge: Found vert[%v]=%v on face[%v]=%+v!!!", pIdx, fi.m.Verts[pIdx], faceIdx, face)
				result = append(result, &vertOnEdgeInfoT{
					vertOnEdgeIdx: pIdx,
					fromVertIdx:   vertIdx,
				})
			}
		}
	}

	return result
}

func genPOnP1Func(p1 Vec3) func(p Vec3) bool {
	switch {
	case p1.X != 0 && p1.Y != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			ty := p.Y / p1.Y
			tz := p.Z / p1.Z
			return tx > 0 && tx < 1 && ty > 0 && ty < 1 && tz > 0 && tz < 1 && AboutEq(tx, ty) && AboutEq(ty, tz)
			// log.Printf("A: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.X != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			tz := p.Z / p1.Z
			return tx > 0 && tx < 1 && tz > 0 && tz < 1 && AboutEq(p.Y, 0) && AboutEq(tx, tz)
			// log.Printf("B: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.X != 0 && p1.Y != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			ty := p.Y / p1.Y
			return tx > 0 && tx < 1 && ty > 0 && ty < 1 && AboutEq(p.Z, 0) && AboutEq(tx, ty)
			// log.Printf("C: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.Y != 0 && p1.Z != 0:
		return func(p Vec3) bool {
			ty := p.Y / p1.Y
			tz := p.Z / p1.Z
			return ty > 0 && ty < 1 && tz > 0 && tz < 1 && AboutEq(p.X, 0) && AboutEq(ty, tz)
			// log.Printf("D: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.X != 0:
		return func(p Vec3) bool {
			tx := p.X / p1.X
			return tx > 0 && tx < 1 && AboutEq(p.Y, 0) && AboutEq(p.Z, 0)
			// log.Printf("E: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.Y != 0:
		return func(p Vec3) bool {
			ty := p.Y / p1.Y
			return ty > 0 && ty < 1 && AboutEq(p.X, 0) && AboutEq(p.Z, 0)
			// log.Printf("F: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	case p1.Z != 0:
		return func(p Vec3) bool {
			tz := p.Z / p1.Z
			return tz > 0 && tz < 1 && AboutEq(p.X, 0) && AboutEq(p.Y, 0)
			// log.Printf("G: pOnP1(p1=%v): p=%v, v=%v", p1, p, v)
			// return v
		}
	default:
		log.Fatalf("programming error: p1=%v", p1)
	}
	return nil
}

/*
decimatePhase1: vertIdx=2 {0.50 -0.50 4.50}, faceIdxes=[0 3 5 6 9 10]
face[0]={{-0.50 -0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50} {-0.50 -0.50 4.50}}
face[3]={{-0.50 -0.50 4.50} {0.50 -0.50 4.50} {0.50 1.50 4.50} {-0.50 1.50 4.50}}
face[5]={{0.50 1.50 4.50} {0.50 -0.50 4.50} {0.50 -0.50 3.50} {0.50 1.50 3.50}}
face[6]={{0.50 -0.50 3.50} {2.50 -0.50 3.50} {2.50 -0.50 4.50} {0.50 -0.50 4.50}}
face[9]={{0.50 -0.50 4.50} {2.50 -0.50 4.50} {2.50 0.50 4.50} {0.50 0.50 4.50}}
face[10]={{0.50 0.50 4.50} {0.50 0.50 3.50} {0.50 -0.50 3.50} {0.50 -0.50 4.50}}
faceNormal[0]={0.00000 -1.00000 0.00000}
faceNormal[3]={-0.00000 0.00000 1.00000}
faceNormal[5]={1.00000 0.00000 -0.00000}
faceNormal[6]={0.00000 -1.00000 0.00000}
faceNormal[9]={-0.00000 0.00000 1.00000}
faceNormal[10]={-1.00000 0.00000 0.00000}
fromVertIdx=2, nonManiVerts[0] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 6, length=2, onFaces: [3 5]}, (toVert={0.50000 1.50000 4.50000})
fromVertIdx=2, nonManiVerts[1] = {n: {0.00000 1.00000 0.00000}, toVertIdx: 11, length=1, onFaces: [9 10]}, (toVert={0.50000 0.50000 4.50000})
nonManiVerts: got 2
cutFaceUsing: cutFaceIdx=3, cuttingVertIdx=11, cuttingFaces=[9 10]

splitOppositeFaceEdge: cutFaceIdx=3, cuttingVertIdx=11, fromVertIdx=2, toVertIdx=6
cutVector from vert[11]={0.50000 0.50000 4.50000}: {-1.00000 0.00000 -0.00000}
found opposite edge: i=3, p1=5={-0.50000 1.50000 4.50000}: lhs={0.00000 0.00000 -2.00000}, rhs={0.00000 0.00000 -1.00000}, ratio=0.5
splitting face: cutInfo=&{5 {-0.5 0.5 4.5} 3}

cutFaceUsing: cutFaceIdx=5, cuttingVertIdx=11, cuttingFaces=[9 10]
Found vert[10]={0.50000 0.50000 3.50000} on face[5]=[6 2 1 7]!!!
Found cutting vert: verts[10]={0.50000 0.50000 3.50000}
Found cutting verts: [10]

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
