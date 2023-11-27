package nodes

import (
	"fmt"
	"log"
	"math"
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

// faceStrToFaceIdxMapT maps a face "signature" (e.g. "0 1 2 3") to a face index.
type faceStrToFaceIdxMapT map[faceKeyT]faceIndexT

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
	faceInfo         *faceInfoT
	faces            []FaceT
	faceNormals      []Vec3
	vertToFaces      vertToFacesMapT
	edgeToFaces      edgeToFacesMapT
	faceStrToFaceIdx faceStrToFaceIdxMapT
	badEdges         edgeToFacesMapT
	badFaces         face2EdgesMapT

	facesTargetedForDeletion map[faceIndexT]bool
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

func makeFaceKeyFromEdges(edges []edgeT) faceKeyT {
	verts := map[VertIndexT]struct{}{}
	for _, edge := range edges {
		verts[edge[0]] = struct{}{}
		verts[edge[1]] = struct{}{}
	}

	face := FaceT(maps.Keys(verts))
	return face.toKey()
}

// genFaceInfo calculates the face normals for every src and dst face
// and generates a map of good and bad edges (mapped to their respective faces).
func (m *Mesh) genFaceInfo(dstFaces, srcFaces []FaceT) *faceInfoT {
	fi := &faceInfoT{m: m}
	fi.src = fi.genFaceInfoForSet(srcFaces)
	fi.dst = fi.genFaceInfoForSet(dstFaces)
	return fi
}

// regenerateFaceInfo regenerates the face info and returns a new struct.
func regenerateFaceInfo(fi *faceInfoT) *faceInfoT {
	newFI := fi.m.genFaceInfo(fi.dst.faces, fi.src.faces)
	newFI.src.facesTargetedForDeletion = fi.src.facesTargetedForDeletion
	newFI.dst.facesTargetedForDeletion = fi.dst.facesTargetedForDeletion
	return newFI
}

func (fi *faceInfoT) genFaceInfoForSet(faces []FaceT) *infoSetT {
	infoSet := &infoSetT{
		faceInfo:         fi,
		faces:            faces,
		faceNormals:      make([]Vec3, 0, len(faces)),
		vertToFaces:      vertToFacesMapT{}, // key=vertIdx, value=[]faceIdx
		edgeToFaces:      edgeToFacesMapT{},
		faceStrToFaceIdx: faceStrToFaceIdxMapT{},
		badEdges:         edgeToFacesMapT{},
		badFaces:         face2EdgesMapT{},

		facesTargetedForDeletion: map[faceIndexT]bool{},
	}

	for i, face := range faces {
		faceIdx := faceIndexT(i)
		infoSet.faceNormals = append(infoSet.faceNormals, fi.m.CalcFaceNormal(face))
		infoSet.faceStrToFaceIdx[face.toKey()] = faceIdx
		for j, vertIdx := range face {
			infoSet.vertToFaces[vertIdx] = append(infoSet.vertToFaces[vertIdx], faceIdx)
			nextVertIdx := face[(j+1)%len(face)]
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
	for faceStr, dstFaceIdx := range fi.dst.faceStrToFaceIdx {
		if srcFaceIdx, ok := fi.src.faceStrToFaceIdx[faceStr]; ok {
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

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on `edge`)
// and therefore is completely independent of the winding order of the face!
// Both edges are found in the `badEdges` map.
// In addition to the edge vector, it also returns the VertIndexT of the other vertex.
func (is *infoSetT) connectedBadEdgeVectorFromVert(vertIdx VertIndexT, edge edgeT) edgeVectorT {
	notVertIdx := edge[0]
	if notVertIdx == vertIdx {
		notVertIdx = edge[1]
	}

	for otherEdge := range is.badEdges {
		var nextIdx VertIndexT
		switch {
		case otherEdge[0] == vertIdx && otherEdge[1] != notVertIdx:
			nextIdx = otherEdge[1]
		case otherEdge[1] == vertIdx && otherEdge[0] != notVertIdx:
			nextIdx = otherEdge[0]
		default:
			continue
		}

		return is.faceInfo.m.makeEdgeVector(vertIdx, nextIdx)
	}

	log.Fatalf("connectedBadEdgeVectorFromVert: programming error for edge %v", edge)
	return edgeVectorT{}
}

func (m *Mesh) makeEdgeVector(fromIdx, toIdx VertIndexT) edgeVectorT {
	toSubFrom := m.Verts[toIdx].Sub(m.Verts[fromIdx])
	return edgeVectorT{
		edge:        makeEdge(fromIdx, toIdx),
		fromVertIdx: fromIdx,
		toVertIdx:   toIdx,
		toSubFrom:   toSubFrom,
		length:      toSubFrom.Length(),
	}
}

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on `edge`)
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

// moveVerts creates new (or reuses old) vertices and returns the mapping from the
// old face's vertIndexes to the new vertices, without modifying the face.
func (is *infoSetT) moveVerts(face FaceT, move Vec3) vToVMap {
	m := is.faceInfo.m

	vertsOldToNew := make(vToVMap, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx].Add(move)
		newVertIdx := m.AddVert(v)
		vertsOldToNew[vertIdx] = newVertIdx
	}

	return vertsOldToNew
}

type vToVMap map[VertIndexT]VertIndexT
type faceSetT map[faceIndexT]struct{}

// moveVertsAlongEdgeLoop creates new (or reuses old) vertices and returns the mapping from the
// old face's vertIndexes to the new vertices, without modifying the face. It moves the
// vertices a given amount along each _connected_ edge (not along the baseFaceIdx edge).
// Note that this must operate not only on this base face, but also on every face that
// shares two edges with this one.
// Also note that each edge vector in edgeLoop points along the direction of movement.
func (is *infoSetT) moveVertsAlongEdgeLoop(baseFaceIdx faceIndexT, amount float64) (vToVMap, faceSetT) {
	log.Printf("moveVertsAlongEdgeLoop(baseFaceIdx=%v, amount=%v)", baseFaceIdx, amount)
	edgeLoop, shortenedFaces := is.findEdgeLoop(baseFaceIdx)

	m := is.faceInfo.m
	result := vToVMap{}
	evsByFace := map[faceIndexT][]edgeVectorT{}
	for _, ev := range edgeLoop {
		is.moveEdge(ev, amount, result)

		for _, faceIdx := range is.edgeToFaces[ev.edge] {
			if _, ok := shortenedFaces[faceIdx]; !ok {
				continue
			}
			evsByFace[faceIdx] = append(evsByFace[faceIdx], ev)
		}
	}

	// now move all verts along edge loops for faces that have more than 4 verts each.
	for faceIdx := range shortenedFaces {
		face := is.faces[faceIdx]
		if len(face) <= 4 {
			continue
		}
		evs := evsByFace[faceIdx]
		log.Printf("moveVertsAlongEdgeLoop: processing face %v", m.dumpFace(faceIdx, face))
		for i, ev := range evs {
			log.Printf("moveVertsAlongEdgeLoop: face evs #%v of %v: %v", i+1, len(evs), ev)
		}
		if len(evs) != 2 {
			log.Printf("WARNING: moveVertsAlongEdgeLoop: unhandled case len(evs)=%v, want 2", len(evs))
			continue
		}

		is.lerpMoveEdgesAlongFace(faceIdx, amount, evs[0], evs[1], result)
	}

	return result, shortenedFaces
}

func (is *infoSetT) findEdgeLoop(baseFaceIdx faceIndexT) (map[edgeT]edgeVectorT, faceSetT) {
	m := is.faceInfo.m
	result := map[edgeT]edgeVectorT{}
	affectedFaces := faceSetT{}
	baseFace := is.faces[baseFaceIdx]

	// first pass - only process vertices that share 3 connected faces to get the prevailing edge vector
	for i, vertIdx := range baseFace {
		nextIdx := baseFace[(i+1)%len(baseFace)]
		edge := makeEdge(vertIdx, nextIdx)

		connectedFaces, ok := is.vertToFaces[vertIdx]
		if !ok {
			log.Fatalf("programming error: findEdgeLoop: no connected faces to vertIdx=%v on edge %v", vertIdx, edge)
		}

		if len(connectedFaces) != 3 { // save for second pass
			continue
		}

		for _, fIdx := range connectedFaces {
			if _, ok := affectedFaces[fIdx]; ok || fIdx == baseFaceIdx {
				continue
			}
			sharedEdge := is.findSharedEdge(baseFaceIdx, fIdx)
			log.Printf("findEdgeLoop: common case: checking out face %v", m.dumpFace(fIdx, is.faces[fIdx]))
			ev := is.connectedEdgeVectorFromVertOnFace(sharedEdge[0], sharedEdge, fIdx)
			log.Printf("findEdgeLoop: common case: from sharedEdge %v and vertIdx=%v: found ev #%v: %v on faceIdx=%v", sharedEdge, sharedEdge[0], len(result)+1, ev, fIdx)
			affectedFaces[fIdx] = struct{}{}
			result[ev.edge] = ev
			for _, otherFaceOnEdge := range is.edgeToFaces[ev.edge] {
				if _, ok := affectedFaces[otherFaceOnEdge]; ok || otherFaceOnEdge == fIdx {
					continue
				}
				log.Printf("findEdgeLoop: common case: from sharedEdge %v and vertIdx=%v: adding other affected face %v", sharedEdge, sharedEdge[0], m.dumpFace(otherFaceOnEdge, is.faces[otherFaceOnEdge]))
				affectedFaces[otherFaceOnEdge] = struct{}{}
			}

			ev = is.connectedEdgeVectorFromVertOnFace(sharedEdge[1], sharedEdge, fIdx) // add other edge
			log.Printf("findEdgeLoop: common case: from sharedEdge %v and vertIdx=%v: found ev #%v: %v on faceIdx=%v", sharedEdge, sharedEdge[1], len(result)+1, ev, fIdx)
			result[ev.edge] = ev
			for _, otherFaceOnEdge := range is.edgeToFaces[ev.edge] {
				if _, ok := affectedFaces[otherFaceOnEdge]; ok || otherFaceOnEdge == fIdx {
					continue
				}
				log.Printf("findEdgeLoop: common case: from sharedEdge %v and vertIdx=%v: adding other affected face %v", sharedEdge, sharedEdge[1], m.dumpFace(otherFaceOnEdge, is.faces[otherFaceOnEdge]))
				affectedFaces[otherFaceOnEdge] = struct{}{}
			}
		}
	}

	if len(result) == 0 {
		log.Printf("WARNING: findEdgeLoop: unhandled case no common cases found")
		return result, affectedFaces
	}

	var refEV edgeVectorT
	for _, v := range result {
		refEV = v
		break
	}
	log.Printf("findEdgeLoop: refEV=%v", refEV)

	// second pass - only process vertices that share 4 connected faces to find extended edge loops
	for i, vertIdx := range baseFace {
		nextIdx := baseFace[(i+1)%len(baseFace)]
		edge := makeEdge(vertIdx, nextIdx)
		log.Printf("findEdgeLoop: second pass: vertIdx=%v, edge=%v", vertIdx, edge)
		if _, ok := result[edge]; ok {
			log.Printf("findEdgeLoop: second pass: skipping edge %v as already processed", edge)
			continue // already processed edge by looping around faces below
		}

		connectedFaces, ok := is.vertToFaces[vertIdx]
		if !ok {
			log.Fatalf("programming error: findEdgeLoop: no connected faces to vertIdx=%v on edge %v", vertIdx, edge)
		}

		if len(connectedFaces) == 3 {
			for _, fIdx := range connectedFaces {
				if _, ok := affectedFaces[fIdx]; ok || fIdx == baseFaceIdx {
					continue
				}
				log.Printf("findEdgeLoop: adding affectedFaces: %v", m.dumpFace(fIdx, is.faces[fIdx]))
				affectedFaces[fIdx] = struct{}{}
			}
			continue
		}

		for _, otherFaceIdx := range connectedFaces {
			if otherFaceIdx == baseFaceIdx {
				continue
			}
			log.Printf("findEdgeLoop: second pass: checking out face %v", m.dumpFace(otherFaceIdx, is.faces[otherFaceIdx]))
			ev := is.connectedEdgeVectorFromVertOnFace(vertIdx, edge, otherFaceIdx)
			if _, ok := result[ev.edge]; !ok {
				dotProduct := Vec3Dot(refEV.toSubFrom, ev.toSubFrom)
				if AboutEq(dotProduct, 0) {
					continue
				}
				log.Printf("findEdgeLoop: face #%v of 4: %v", i+1, m.dumpFace(otherFaceIdx, is.faces[otherFaceIdx]))
				log.Printf("findEdgeLoop: dotProduct=%v, found ev #%v: %v on faceIdx %v", dotProduct, len(result)+1, ev, otherFaceIdx)
				affectedFaces[otherFaceIdx] = struct{}{}
				result[ev.edge] = ev
			}

			// Now trace this edge all the way around until a known edge is found.
			// This currently assumes that all these faces are quads.
			is.followQuadFacesEdgeLoop(result, affectedFaces, ev.edge, otherFaceIdx)
		}
	}
	return result, affectedFaces
}

func (is *infoSetT) findSharedEdge(f1Idx, f2Idx faceIndexT) edgeT {
	if len(is.faces[f1Idx]) > len(is.faces[f2Idx]) {
		f1Idx, f2Idx = f2Idx, f1Idx
	}
	edges := map[edgeT]struct{}{}
	face1 := is.faces[f1Idx]
	for i, vertIdx := range face1 {
		nextIdx := face1[(i+1)%len(face1)]
		edge := makeEdge(vertIdx, nextIdx)
		edges[edge] = struct{}{}
	}
	face2 := is.faces[f2Idx]
	for i, vertIdx := range face2 {
		nextIdx := face2[(i+1)%len(face2)]
		edge := makeEdge(vertIdx, nextIdx)
		if _, ok := edges[edge]; ok {
			return edge
		}
	}
	// log.Printf("findSharedEdge: programming error: no shared edges between face index %v and %v", f1Idx, f2Idx)
	// m := is.faceInfo.m
	// log.Printf("findSharedEdge: %v", m.dumpFace(f1Idx, is.faces[f1Idx]))
	// log.Fatalf("findSharedEdge: %v", m.dumpFace(f2Idx, is.faces[f2Idx]))
	return edgeT{}
}

func (is *infoSetT) followQuadFacesEdgeLoop(result map[edgeT]edgeVectorT, affectedFaces faceSetT, knownEdge edgeT, faceIdx faceIndexT) {
	m := is.faceInfo.m
	for {
		log.Printf("followQuadFacesEdgeLoop: knownEdge=%v, face: %v", knownEdge, m.dumpFace(faceIdx, is.faces[faceIdx]))
		otherEdgeVector, ok := is.otherQuadEdge(knownEdge, faceIdx)
		if !ok {
			log.Printf("WARNING: unhandled case: followQuadFacesEdgeLoop: face %v could not find other edge from %v", faceIdx, knownEdge)
			return
		}

		knownEdge = otherEdgeVector.edge
		if _, ok := result[knownEdge]; ok { // done with the loop
			log.Printf("followQuadFacesEdgeLoop: ending loop because already processed edge %v - but did we get both vertices of that edge?!?", knownEdge)
			return
		}
		log.Printf("followQuadFacesEdgeLoop: found next ev #%v: %v on faceIdx %v", len(result)+1, otherEdgeVector, faceIdx)
		affectedFaces[faceIdx] = struct{}{}
		result[knownEdge] = otherEdgeVector

		connectedFaces, ok := is.edgeToFaces[knownEdge]
		if !ok || len(connectedFaces) != 2 {
			log.Printf("WARNING! followQuadFacesEdgeLoop: expected 2 faces on edge %v, got %v", knownEdge, len(connectedFaces))
			return
		}

		nextFace := connectedFaces[0]
		if nextFace == faceIdx {
			nextFace = connectedFaces[1]
		}
		faceIdx = nextFace
	}
}

// otherQuadEdge makes a new edge vector that is facing the SAME DIRECTION as the reference
// edge (which always simply reverses the direction of the opposite edge vector).
// It is possible that the face is not a quad. So to find the opposing edge, we find
// the opposite edge whose edge unit vertex normal points in the opposite direction of the reference edge.
func (is *infoSetT) otherQuadEdge(refEdge edgeT, faceIdx faceIndexT) (edgeVectorT, bool) {
	face := is.faces[faceIdx]
	// first, find this edge in the face in order to get the edge unit vector.
	var refUV Vec3
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		edge := makeEdge(vertIdx, nextIdx)
		if edge == refEdge {
			refUV = is.faceInfo.m.makeEdgeVector(vertIdx, nextIdx).toSubFrom.Normalized()
			// log.Printf("otherQuadEdge: refUV=%v", refUV)
			break
		}
	}

	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if vertIdx == refEdge[0] || vertIdx == refEdge[1] || nextIdx == refEdge[0] || nextIdx == refEdge[1] {
			continue
		}
		negatedEV := is.faceInfo.m.makeEdgeVector(nextIdx, vertIdx) // VERTEX ORDER SWAPPED ON PURPOSE HERE!!!
		uv := negatedEV.toSubFrom.Normalized()
		if uv.AboutEq(refUV) {
			return negatedEV, true
		}
	}
	log.Printf("WARNING: otherQuadEdge: unable to find opposite edge for %v: %v", refEdge, refUV)
	return edgeVectorT{}, false
}

func (is *infoSetT) lerpMoveEdgesAlongFace(faceIdx faceIndexT, amount float64, ev1, ev2 edgeVectorT, result vToVMap) {
	log.Printf("lerpMoveEdgesAlongFace(faceIdx=%v, amount=%v, ev1=%+v, ev2=%+v", faceIdx, amount, ev1, ev2)
	vertsBetweenEdges := is.vertsBetweenEdges(faceIdx, ev1, ev2)
	if len(vertsBetweenEdges) == 0 {
		return
	}
	log.Printf("lerpMoveEdgesAlongFace: vertsBetweenEdges=%+v", vertsBetweenEdges)

	m := is.faceInfo.m
	uv1 := ev1.toSubFrom.Normalized()
	uv2 := ev2.toSubFrom.Normalized()
	for i, vertIdx := range vertsBetweenEdges {
		t := float64(i+1) / float64(len(vertsBetweenEdges)+1)
		t1 := 1 - t
		lerp := Vec3Add(uv1.MulScalar(t1), uv2.MulScalar(t))
		move := lerp.MulScalar(amount)

		v := m.Verts[vertIdx].Add(move)
		newVertIdx := m.AddVert(v)
		log.Printf("lerpMoveEdgeAlongFace: from old vert %v=%v, creating new vert %v at %v", vertIdx, m.Verts[vertIdx].toKey(), newVertIdx, v.toKey())
		result[vertIdx] = newVertIdx
		log.Printf("lerpMoveEdgeAlongFace: ev1=%v, ev2=%v, move=%v, oldVert[%v]=%v, newVert[%v]=%v",
			ev1, ev2, move, vertIdx, m.Verts[vertIdx], newVertIdx, m.Verts[newVertIdx])
	}
}

// Note that vertsBetweenEdges returns the verts in order from ev1 to ev2 so that they
// can be linearly interpolated properly above.
func (is *infoSetT) vertsBetweenEdges(faceIdx faceIndexT, ev1, ev2 edgeVectorT) []VertIndexT {
	face := is.faces[faceIdx]
	startI, endI := -1, -1
	var startIVertIdx VertIndexT
	for i, vIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if (vIdx == ev1.fromVertIdx && nextIdx == ev2.fromVertIdx) ||
			(vIdx == ev2.fromVertIdx && nextIdx == ev1.fromVertIdx) {
			return nil // no intervening vertices
		}
		if (vIdx == ev1.fromVertIdx && nextIdx == ev1.toVertIdx) ||
			(vIdx == ev2.fromVertIdx && nextIdx == ev2.toVertIdx) {
			endI = (i - 1 + len(face)) % len(face) // inclusive
		}
		if vIdx == ev1.toVertIdx && nextIdx == ev1.fromVertIdx {
			startIVertIdx = ev1.fromVertIdx
			startI = (i + 2) % len(face) // inclusive
		} else if vIdx == ev2.toVertIdx && nextIdx == ev2.fromVertIdx {
			startIVertIdx = ev2.fromVertIdx
			startI = (i + 2) % len(face) // inclusive
		}
	}

	if startI < 0 || endI < 0 {
		log.Printf("WARNING: vertsBetweenEdges: programming error: startI=%v, endI=%v, want >= 0", startI, endI)
	}
	if startI > endI {
		startI, endI = endI, startI
	}

	result := make([]VertIndexT, 0, endI-startI+1)
	for i := startI; i <= endI; i++ {
		result = append(result, face[i])
	}
	if startIVertIdx == ev2.fromVertIdx {
		slices.Reverse(result)
	}
	return result
}

func (is *infoSetT) moveEdge(ev edgeVectorT, amount float64, vertsOldToNew vToVMap) {
	uv := ev.toSubFrom.Normalized()
	move := uv.MulScalar(amount)

	m := is.faceInfo.m
	v := m.Verts[ev.fromVertIdx].Add(move)
	newVertIdx := m.AddVert(v)
	// log.Printf("moveEdge: from old vert %v=%v, creating new vert %v at %v", ev.fromVertIdx, m.Verts[ev.fromVertIdx].toKey(), newVertIdx, v.toKey())
	vertsOldToNew[ev.fromVertIdx] = newVertIdx
	// log.Printf("moveEdge: ev=%v, uv=%v, move=%v, oldVert[%v]=%v, newVert[%v]=%v",
	// 	ev, uv, move, ev.fromVertIdx, m.Verts[ev.fromVertIdx], newVertIdx, m.Verts[newVertIdx])
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
		return v1 * v2
	}

	p1 := m.Verts[face[0]]
	p2 := m.Verts[face[1]]
	p3 := m.Verts[face[2]]
	a := math.Pow(((p2.Y-p1.Y)*(p3.Z-p1.Z)-(p3.Y-p1.Y)*(p2.Z-p1.Z)), 2) + math.Pow(((p3.X-p1.X)*(p2.Z-p1.Z)-(p2.X-p1.X)*(p3.Z-p1.Z)), 2) + math.Pow(((p2.X-p1.X)*(p3.Y-p1.Y)-(p3.X-p1.X)*(p2.Y-p1.Y)), 2)
	cosnx := ((p2.Y-p1.Y)*(p3.Z-p1.Z) - (p3.Y-p1.Y)*(p2.Z-p1.Z)) / math.Sqrt(a)
	cosny := ((p3.X-p1.X)*(p2.Z-p1.Z) - (p2.X-p1.X)*(p3.Z-p1.Z)) / math.Sqrt(a)
	cosnz := ((p2.X-p1.X)*(p3.Y-p1.Y) - (p3.X-p1.X)*(p2.Y-p1.Y)) / math.Sqrt(a)
	var s float64
	for i, vertIdx := range face {
		p1 = m.Verts[vertIdx]
		p2 = m.Verts[face[(i+1)%len(face)]]
		s += cosnz*((p1.X)*(p2.Y)-(p2.X)*(p1.Y)) + cosnx*((p1.Y)*(p2.Z)-(p2.Y)*(p1.Z)) + cosny*((p1.Z)*(p2.X)-(p2.Z)*(p1.X))
	}

	return math.Abs(0.5 * s)
}

func (m *Mesh) dumpFaces(faces []FaceT) string {
	var lines []string
	for i, face := range faces {
		lines = append(lines, m.dumpFace(faceIndexT(i), face))
	}
	return strings.Join(lines, "\n")
}

func (is *infoSetT) dumpFaceIndices(faceIdxes []faceIndexT) string {
	var lines []string
	for _, faceIdx := range faceIdxes {
		face := is.faces[faceIdx]
		lines = append(lines, is.faceInfo.m.dumpFace(faceIdx, face))
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
