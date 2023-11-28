// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"
)

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh (ideally, although it is possible that it is still non-manifold).
func (fi *faceInfoT) mergeNonManifoldSrc() {
	edgeLoops := fi.src.badEdgesToConnectedEdgeLoops()
	// log.Printf("mergeNonManifoldSrc: src:\n%v", fi.m.dumpFaces(fi.src.faces))
	// log.Printf("mergeNonManifoldSrc: dst:\n%v", fi.m.dumpFaces(fi.dst.faces))
	// log.Printf("mergeNonManifoldSrc: edgeLoops: %+v", edgeLoops)

cutsMade:
	for faceStr, edges := range edgeLoops {
		if deleteFaceIdx, ok := fi.dst.faceStrToFaceIdx[faceStr]; ok {
			// log.Printf("mergeNonManifoldSrc: faceStr found in dst: %v, deleting face: %v", faceStr, deleteFaceIdx)
			fi.dst.facesTargetedForDeletion[deleteFaceIdx] = true
			continue
		}

		// log.Printf("mergeNonManifoldSrc: faceStr not found in dst: %v", faceStr)
		// log.Printf("mergeNonManifoldSrc: src.badEdges: %+v", fi.src.badEdges)
		// log.Printf("mergeNonManifoldSrc: dst.edgeToFaces: %+v", fi.dst.edgeToFaces)

		// Find a dst face that shares two (not joined) edge unit vectors with this srcFace,
		// then resize it accordingly.
		for _, edge := range edges {
			srcFaceIndices, ok := fi.src.badEdges[edge]
			if !ok || len(srcFaceIndices) != 1 {
				// this is not a valid edge connected to a singular face so keep looking
				continue
			}
			// srcFaceIdx := srcFaceIndices[0] // This is the only src face that shares an edge with a dst face.
			// log.Printf("Looking at shared edge: %v from src %v", edge, fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
			srcE1EV := fi.src.connectedBadEdgeVectorFromVert(edge[0], edge)
			srcE1UV := srcE1EV.toSubFrom.Normalized()
			srcE2EV := fi.src.connectedBadEdgeVectorFromVert(edge[1], edge)
			srcE2UV := srcE2EV.toSubFrom.Normalized()

			// log.Printf("mergeNonManifoldSrc: single-cut srcE1EV=%+v", srcE1EV)
			// log.Printf("mergeNonManifoldSrc: single-cut srcE1UV=%+v", srcE1UV)
			// log.Printf("mergeNonManifoldSrc: single-cut srcE2EV=%+v", srcE2EV)
			// log.Printf("mergeNonManifoldSrc: single-cut srcE2UV=%+v", srcE2UV)

			dstFaceIdx, dstEVs, ok := fi.dst.findFaceSharingTwoEdgeUVs(edge, srcE1UV, srcE2UV)
			if !ok {
				continue
			}

			// log.Printf("mergeNonManifoldSrc: single cutting neighbors of dstFaceIdx: %v: %+v", dstFaceIdx, fi.dst.faces[dstFaceIdx])
			srcEVs := [2]edgeVectorT{srcE1EV, srcE2EV}
			fi.dst.resizeFace(nil, dstFaceIdx, dstEVs[0].edge, dstEVs[1].edge, srcEVs) // resize dst by shorter edge vectors
			continue cutsMade
		}

		// No cuts were made at this point. Check if any dst face shares the same vertex with two edges on
		// the open edge loop. If so, insert a vertex into the enclosing face to walk around the src face.
		corners := edgeLoopCorners(edges)
		for vertIdx, otherVerts := range corners { // no srcFaceIdx because it is an open edge loop
			// log.Printf("mergeNonManifoldSrc: double-cut looking at vertIdx=%v and otherVerts=%+v", vertIdx, otherVerts)
			srcE1EV := fi.m.makeEdgeVector(vertIdx, otherVerts[0])
			srcE1UV := srcE1EV.toSubFrom.Normalized()
			// log.Printf("mergeNonManifoldSrc: double-cut srcE1EV=%+v", srcE1EV)
			// log.Printf("mergeNonManifoldSrc: double-cut srcE1UV=%+v", srcE1UV)

			srcE2EV := fi.m.makeEdgeVector(vertIdx, otherVerts[1])
			srcE2UV := srcE2EV.toSubFrom.Normalized()
			// log.Printf("mergeNonManifoldSrc: double-cut srcE2EV=%+v", srcE2EV)
			// log.Printf("mergeNonManifoldSrc: double-cut srcE2UV=%+v", srcE2UV)

			dstFaceIdx, dstEVs, ok := fi.dst.findFaceSharingTwoEdgeUVsFromVert(vertIdx, srcE1UV, srcE2UV)
			if !ok {
				continue
			}

			// log.Printf("mergeNonManifoldSrc: double-cut found dst face sharing two edges from vertIdx=%v: %v", vertIdx, fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
			// log.Printf("mergeNonManifoldSrc: dstEVs[0]=%v", dstEVs[0])
			// log.Printf("mergeNonManifoldSrc: dstEVs[1]=%v", dstEVs[1])

			cornerNormal := Vec3Cross(srcE1UV, srcE2UV)
			// log.Printf("mergeNonManifoldSrc: corner normal: %v", cornerNormal)
			dstFaceNormal := fi.dst.faceNormals[dstFaceIdx]
			// log.Printf("mergeNonManifoldSrc: dst face normal: %v", dstFaceNormal)

			// now replace the corner with two new vertices, then fill in the leftover src vertices between those two.
			if !cornerNormal.AboutEq(dstFaceNormal) {
				// opposite winding order - switch the corner verts
				otherVerts[0], otherVerts[1] = otherVerts[1], otherVerts[0]
			}

			dstFace := fi.dst.faces[dstFaceIdx]
			dstCornerI := slices.Index(dstFace, vertIdx)
			if dstCornerI < 0 {
				log.Fatalf("programming error: dstCornerI=%v, want >=0", dstCornerI)
			}
			// first, remove the corner vertex
			dstFace = slices.Delete(dstFace, dstCornerI, dstCornerI+1)
			// next, insert the two src corner verts
			dstFace = slices.Insert(dstFace, dstCornerI, otherVerts[1], otherVerts[0]) // order is important
			// now, add the remaining verts from the edge loop
			verts := remainingVerts(vertIdx, otherVerts[1], otherVerts[0], corners)
			dstFace = slices.Insert(dstFace, dstCornerI+1, verts...)
			// finally, replace the face
			fi.dst.faces[dstFaceIdx] = dstFace

			// Now, clean up all other edges affected by the changing of dstFace
			fi.dst.addVertToEdge(dstEVs[0].edge, otherVerts[1])
			fi.dst.addVertToEdge(dstEVs[1].edge, otherVerts[0])

			continue cutsMade
		}
	}
}

func (is *infoSetT) addVertToEdge(edge edgeT, vertIdx VertIndexT) {
	for _, faceIdx := range is.edgeToFaces[edge] {
		is.addVertToFaceEdge(faceIdx, edge, vertIdx)
	}
}

func (is *infoSetT) addVertToFaceEdge(faceIdx faceIndexT, edge edgeT, vertIdx VertIndexT) {
	face := is.faces[faceIdx]
	for i, vIdx := range face {
		nextI := (i + 1) % len(face)
		nextIdx := face[nextI]
		if makeEdge(vIdx, nextIdx) != edge {
			continue
		}
		is.faces[faceIdx] = slices.Insert(face, nextI, vertIdx)
		// log.Printf("addVertToFaceEdge(faceIdx=%v, edge=%v): inserting vertIdx=%v at position %v", faceIdx, edge, vertIdx, nextI)
		// log.Printf("addVertToFaceEdge: result: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
	}
}

func remainingVerts(cornerVertIdx, startVertIdx, otherVertIdx VertIndexT, corners cornerT) []VertIndexT {
	seen := map[VertIndexT]bool{
		cornerVertIdx: true,
		startVertIdx:  true,
		otherVertIdx:  true,
	}
	result := make([]VertIndexT, 0, len(corners)-3)
	for i := 0; i < len(corners)-3; i++ {
		corner := corners[startVertIdx]
		if seen[corner[0]] {
			result = append(result, corner[1])
			seen[corner[1]] = true
			continue
		}
		result = append(result, corner[0])
		seen[corner[0]] = true
	}
	return result
}

func (is *infoSetT) findFaceSharingTwoEdgeUVs(edge edgeT, e1UV, e2UV Vec3) (faceIndexT, [2]edgeVectorT, bool) {
	faceIndices, ok := is.edgeToFaces[edge]
	if !ok {
		return 0, [2]edgeVectorT{}, false
	}

	for _, faceIdx := range faceIndices {
		// log.Printf("Looking at faceIdx: %v: %+v", faceIdx, is.faces[faceIdx])

		evs := is.makeEdgeVectors(edge, faceIdx)
		uvs := []Vec3{evs[0].toSubFrom.Normalized(), evs[1].toSubFrom.Normalized()}
		// log.Printf("evs[0]=%+v", evs[0])
		// log.Printf("uvs[0]=%+v", uvs[0])
		// log.Printf("evs[1]=%+v", evs[1])
		// log.Printf("uvs[1]=%+v", uvs[1])

		if e1UV.AboutEq(uvs[0]) && e2UV.AboutEq(uvs[1]) {
			// log.Printf("Found matching face: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
			return faceIdx, evs, true
		}
	}
	return 0, [2]edgeVectorT{}, false
}

func (is *infoSetT) findFaceSharingTwoEdgeUVsFromVert(vertIdx VertIndexT, e1UV, e2UV Vec3) (faceIndexT, [2]edgeVectorT, bool) {
	faceIndices, ok := is.vertToFaces[vertIdx]
	if !ok {
		return 0, [2]edgeVectorT{}, false
	}

	for _, faceIdx := range faceIndices {
		// log.Printf("Looking at faceIdx: %v: %+v", faceIdx, is.faces[faceIdx])

		evs := is.makeEdgeVectorsFromVert(vertIdx, faceIdx)
		uvs := []Vec3{evs[0].toSubFrom.Normalized(), evs[1].toSubFrom.Normalized()}
		// log.Printf("evs[0]=%+v", evs[0])
		// log.Printf("uvs[0]=%+v", uvs[0])
		// log.Printf("evs[1]=%+v", evs[1])
		// log.Printf("uvs[1]=%+v", uvs[1])

		if (e1UV.AboutEq(uvs[0]) && e2UV.AboutEq(uvs[1])) ||
			(e1UV.AboutEq(uvs[1]) && e2UV.AboutEq(uvs[0])) {
			// log.Printf("Found matching face: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
			return faceIdx, evs, true
		}
	}
	return 0, [2]edgeVectorT{}, false
}

type edgeLoopT struct {
	edges []edgeT
}

func (el *edgeLoopT) addEdge(edge edgeT) {
	for _, v := range el.edges {
		if v == edge {
			return
		}
	}
	el.edges = append(el.edges, edge)
}

func (is *infoSetT) badEdgesToConnectedEdgeLoops() map[faceKeyT][]edgeT {
	vertsToEdgeLoops := map[VertIndexT]*edgeLoopT{}
	edgeLoops := map[*edgeLoopT]*edgeLoopT{}
	newEdgeLoop := func(edge edgeT) {
		el := &edgeLoopT{edges: []edgeT{edge}}
		vertsToEdgeLoops[edge[0]] = el
		vertsToEdgeLoops[edge[1]] = el
		edgeLoops[el] = el
	}
	addEdgeToLoop := func(edge edgeT, el *edgeLoopT) {
		el.addEdge(edge)
		vertsToEdgeLoops[edge[0]] = el
		vertsToEdgeLoops[edge[1]] = el
	}
	mergeTwoEdgeLoopsWithEdge := func(edge edgeT, edgeLoop1, edgeLoop2 *edgeLoopT) {
		addEdgeToLoop(edge, edgeLoop1)
		for _, v := range edgeLoop2.edges {
			edgeLoop1.addEdge(v)
			vertsToEdgeLoops[v[0]] = edgeLoop1
			vertsToEdgeLoops[v[1]] = edgeLoop1
		}
		delete(edgeLoops, edgeLoop2)
	}

	for edge := range is.badEdges {
		edgeLoop1, ok1 := vertsToEdgeLoops[edge[0]]
		edgeLoop2, ok2 := vertsToEdgeLoops[edge[1]]
		switch {
		case ok1 && ok2 && edgeLoop1 == edgeLoop2:
			addEdgeToLoop(edge, edgeLoop1)
		case ok1 && ok2: // && edgeLoop1!=edgeLoop2: - delete the one edge loop and merge into the other
			mergeTwoEdgeLoopsWithEdge(edge, edgeLoop1, edgeLoop2)
		case ok1:
			addEdgeToLoop(edge, edgeLoop1)
		case ok2:
			addEdgeToLoop(edge, edgeLoop2)
		default:
			newEdgeLoop(edge)
		}
	}

	result := make(map[faceKeyT][]edgeT, len(edgeLoops))
	for _, edgeLoop := range edgeLoops {
		key := makeFaceKeyFromEdges(edgeLoop.edges)
		if v, ok := result[key]; ok {
			log.Fatalf("badEdgesToConnectedEdgeLoops: programming error: already assigned faceStr key=%v: old=%+v, new=%+v", key, v, edgeLoop.edges)
		}
		result[key] = edgeLoop.edges
	}

	return result
}

type cornerT map[VertIndexT][]VertIndexT

func edgeLoopCorners(edges []edgeT) cornerT {
	corners := make(cornerT, len(edges))
	addCorner := func(v1, v2 VertIndexT) {
		if vs, ok := corners[v1]; ok {
			corners[v1] = append(vs, v2)
			return
		}
		corners[v1] = []VertIndexT{v2}
	}
	for _, edge := range edges {
		addCorner(edge[0], edge[1])
		addCorner(edge[1], edge[0])
	}
	return corners
}
