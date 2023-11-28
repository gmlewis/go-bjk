// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"
)

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh (ideally, although it is possible that it is still non-manifold).
func (fi *faceInfoT) mergeNonManifoldSrc() {
	// 	// If there are N bad edges and N bad faces, chances are good that these are simply open
	// 	// (unconnected) extrusions looking to join the dst mesh.
	// 	if len(fi.src.badEdges) == len(fi.dst.badFaces) {
	fi.connectOpenSrcExtrusionsToDst()
	// 		return
	// 	}

	/*
		// srcFaceIndicesToEdges := reverseMapBadEdges(fi.src.badEdges)
		// debugFaces := make([]FaceT, 0, len(srcFaceIndicesToEdges))
		// log.Printf("mergeNonManifoldSrc: srcFaceIndicesToEdges: %+v", srcFaceIndicesToEdges)
		// for srcFaceIdx, badEdges := range srcFaceIndicesToEdges {
		// 	debugFaces = append(debugFaces, fi.src.faces[srcFaceIdx])
		// 	log.Printf("mergeNonManifoldSrc: src.faces[%v] has %v bad edges: %+v", srcFaceIdx, len(badEdges), badEdges)
		// }
		// fi.m.Faces = debugFaces
		// fi.m.WriteSTL(fmt.Sprintf("debug-%v-%v-badFaces-%v.stl", len(fi.src.faces), len(fi.dst.faces), len(debugFaces)))

		for edge, faceIdxes := range fi.src.badEdges {
			switch len(faceIdxes) {
			case 1:
				log.Printf("WARNING: mergeNonManifoldSrc: skipping edge %v with one face", edge)
			case 3:
				// debugFileBaseName := fmt.Sprintf("debug-%v-%v-edge-%v-%v-", len(fi.src.faces), len(fi.dst.faces), edge[0], edge[1])
				// log.Printf("debugFileBaseName=%v", debugFileBaseName)
				// fi.m.Faces = fi.src.faces
				// fi.m.WriteSTL(debugFileBaseName + "src.stl")
				// fi.m.Faces = fi.dst.faces
				// fi.m.WriteSTL(debugFileBaseName + "dst.stl")

				fi.src.fixEdge3Faces(edge, faceIdxes)
			default:
				log.Printf("WARNING: mergeNonManifoldSrc: skipping edge %v with %v faces", edge, len(faceIdxes))
			}
		}
	*/
}

func (is *infoSetT) fixEdge3Faces(edge edgeT, faceIdxes []faceIndexT) {
	f0, f1, f2 := faceIdxes[0], faceIdxes[1], faceIdxes[2]
	switch {
	case is.faceNormals[f0].AboutEq(is.faceNormals[f1]):
		is.fixEdge2OverlapingFaces(edge, f0, f1, f2)
	case is.faceNormals[f1].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f1, f2, f0)
	case is.faceNormals[f0].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f0, f2, f1)
	default:
		log.Printf("WARNING: fixEdge3Faces: unhandled case normals: %v %v %v", is.faceNormals[f0], is.faceNormals[f1], is.faceNormals[f2])
	}
}

func (is *infoSetT) fixEdge2OverlapingFaces(edge edgeT, f0, f1, otherFaceIdx faceIndexT) {
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f0=%v: %v", edge, f0, is.faceNormals[f0], is.faceInfo.m.dumpFace(f0, is.faces[f0]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f1=%v: %v", edge, f1, is.faceNormals[f1], is.faceInfo.m.dumpFace(f1, is.faces[f1]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, otherFace=%v: %v", edge, otherFaceIdx, is.faceNormals[otherFaceIdx], is.faceInfo.m.dumpFace(otherFaceIdx, is.faces[otherFaceIdx]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): shared edge: %v - %v", edge, is.faceInfo.m.Verts[edge[0]], is.faceInfo.m.Verts[edge[1]])
	//
	// is.faceInfo.m.Faces = []FaceT{is.faces[f0]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "f0.stl")
	// is.faceInfo.m.Faces = []FaceT{is.faces[f1]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "f1.stl")
	// is.faceInfo.m.Faces = []FaceT{is.faces[otherFaceIdx]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "otherFaceIdx.stl")

	log.Fatalf("fixEdge2OverlapingFaces: STOP")

	// is.facesTargetedForDeletion[otherFaceIdx] = true

	// f0VertKey := is.toVertKey(is.faces[f0])
	// f1VertKey := is.toVertKey(is.faces[f1])
	// if f0VertKey == f1VertKey {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces are identical! f0VertKey=%v", edge, f0VertKey)
	// } else {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces DIFFER!\nf0VertKey=%v\nf1VertKey=%v", edge, f0VertKey, f1VertKey)
	// }
}

func (fi *faceInfoT) connectOpenSrcExtrusionsToDst() {
	edgeLoops := fi.src.badEdgesToConnectedEdgeLoops()
	// log.Printf("connectOpenSrcExtrusionsToDst: src:\n%v", fi.m.dumpFaces(fi.src.faces))
	// log.Printf("connectOpenSrcExtrusionsToDst: dst:\n%v", fi.m.dumpFaces(fi.dst.faces))
	// log.Printf("connectOpenSrcExtrusionsToDst: edgeLoops: %+v", edgeLoops)

cutsMade:
	for faceStr, edges := range edgeLoops {
		if deleteFaceIdx, ok := fi.dst.faceStrToFaceIdx[faceStr]; ok {
			// log.Printf("connectOpenSrcExtrusionsToDst: faceStr found in dst: %v, deleting face: %v", faceStr, deleteFaceIdx)
			fi.dst.facesTargetedForDeletion[deleteFaceIdx] = true
			continue
		}

		log.Printf("connectOpenSrcExtrusionsToDst: faceStr not found in dst: %v", faceStr)
		log.Printf("connectOpenSrcExtrusionsToDst: src.badEdges: %+v", fi.src.badEdges)
		log.Printf("connectOpenSrcExtrusionsToDst: dst.edgeToFaces: %+v", fi.dst.edgeToFaces)

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

			// log.Printf("connectOpenSrcExtrusionsToDst: single-cut srcE1EV=%+v", srcE1EV)
			// log.Printf("connectOpenSrcExtrusionsToDst: single-cut srcE1UV=%+v", srcE1UV)
			// log.Printf("connectOpenSrcExtrusionsToDst: single-cut srcE2EV=%+v", srcE2EV)
			// log.Printf("connectOpenSrcExtrusionsToDst: single-cut srcE2UV=%+v", srcE2UV)

			dstFaceIdx, dstEVs, ok := fi.dst.findFaceSharingTwoEdgeUVs(edge, srcE1UV, srcE2UV)
			if !ok {
				continue
			}

			// log.Printf("connectOpenSrcExtrusionsToDst: single cutting neighbors of dstFaceIdx: %v: %+v", dstFaceIdx, fi.dst.faces[dstFaceIdx])
			srcEVs := [2]edgeVectorT{srcE1EV, srcE2EV}
			fi.dst.resizeFace(nil, dstFaceIdx, dstEVs[0].edge, dstEVs[1].edge, srcEVs) // resize dst by shorter edge vectors
			continue cutsMade
		}

		// No cuts were made at this point. Check if any dst face shares the same vertex with two edges on
		// the open edge loop. If so, insert a vertex into the enclosing face to walk around the src face.
		corners := edgeLoopCorners(edges)
		for vertIdx, otherVerts := range corners { // no srcFaceIdx because it is an open edge loop
			log.Printf("connectOpenSrcExtrusionsToDst: double-cut looking at vertIdx=%v and otherVerts=%+v", vertIdx, otherVerts)
			srcE1EV := fi.m.makeEdgeVector(vertIdx, otherVerts[0])
			srcE1UV := srcE1EV.toSubFrom.Normalized()
			log.Printf("connectOpenSrcExtrusionsToDst: double-cut srcE1EV=%+v", srcE1EV)
			log.Printf("connectOpenSrcExtrusionsToDst: double-cut srcE1UV=%+v", srcE1UV)

			srcE2EV := fi.m.makeEdgeVector(vertIdx, otherVerts[1])
			srcE2UV := srcE2EV.toSubFrom.Normalized()
			log.Printf("connectOpenSrcExtrusionsToDst: double-cut srcE2EV=%+v", srcE2EV)
			log.Printf("connectOpenSrcExtrusionsToDst: double-cut srcE2UV=%+v", srcE2UV)

			dstFaceIdx, dstEVs, ok := fi.dst.findFaceSharingTwoEdgeUVsFromVert(vertIdx, srcE1UV, srcE2UV)
			if !ok {
				continue
			}

			log.Printf("connectOpenSrcExtrusionsToDst: double-cut found dst face sharing two edges from vertIdx=%v: %v", vertIdx, fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
			log.Printf("connectOpenSrcExtrusionsToDst: dstEVs[0]=%v", dstEVs[0])
			log.Printf("connectOpenSrcExtrusionsToDst: dstEVs[1]=%v", dstEVs[1])

			cornerNormal := Vec3Cross(srcE1UV, srcE2UV)
			log.Printf("connectOpenSrcExtrusionsToDst: corner normal: %v", cornerNormal)
			dstFaceNormal := fi.dst.faceNormals[dstFaceIdx]
			log.Printf("connectOpenDstExtrusionsToDst: dst face normal: %v", dstFaceNormal)

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
		log.Printf("addVertToFaceEdge(faceIdx=%v, edge=%v): inserting vertIdx=%v at position %v", faceIdx, edge, vertIdx, nextI)
		log.Printf("addVertToFaceEdge: result: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
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

func (is *infoSetT) otherBadEdgeFromVert(fromVertIdx VertIndexT, refEdge edgeT) edgeT {
	for edge := range is.badEdges {
		if edge != refEdge && (edge[0] == fromVertIdx || edge[1] == fromVertIdx) {
			return edge
		}
	}
	log.Fatalf("otherBadEdgeFromVert: programming error")
	return refEdge
}

func (is *infoSetT) findBaseFaceSharingTwoEdgeUVsFromVert(fromVertIdx VertIndexT, e1UV, e2UV Vec3) (faceIndexT, bool) {
	faceIndices, ok := is.vertToFaces[fromVertIdx]
	if !ok {
		return 0, false
	}

	for _, baseFaceIdx := range faceIndices {
		// log.Printf("Looking at baseFaceIdx: %v: %+v", baseFaceIdx, is.faces[baseFaceIdx])

		face := is.faces[baseFaceIdx]
		var matches int
		for i, vIdx := range face {
			nextIdx := face[(i+1)%len(face)]
			if vIdx != fromVertIdx && nextIdx != fromVertIdx {
				continue
			}
			edge := makeEdge(vIdx, nextIdx)

			myEV := is.connectedEdgeVectorFromVertOnFace(fromVertIdx, edge, baseFaceIdx)
			myUV := myEV.toSubFrom.Normalized()

			if e1UV.AboutEq(myUV) || e2UV.AboutEq(myUV) {
				matches++
			}
		}
		if matches == 2 {
			return baseFaceIdx, true
		}
	}
	return 0, false
}

func (is *infoSetT) findBaseFaceSharingTwoEdgeUVs(edge edgeT, e1UV, e2UV Vec3) (faceIndexT, bool) {
	faceIndices, ok := is.edgeToFaces[edge]
	if !ok {
		return 0, false
	}

	for _, baseFaceIdx := range faceIndices {
		// log.Printf("Looking at baseFaceIdx: %v: %+v", baseFaceIdx, is.faces[baseFaceIdx])

		myE1EV := is.connectedEdgeVectorFromVertOnFace(edge[0], edge, baseFaceIdx)
		myE1UV := myE1EV.toSubFrom.Normalized()
		myE2EV := is.connectedEdgeVectorFromVertOnFace(edge[1], edge, baseFaceIdx)
		myE2UV := myE2EV.toSubFrom.Normalized()
		// log.Printf("myE1EV=%+v", myE1EV)
		// log.Printf("myE1UV=%+v", myE1UV)
		// log.Printf("myE2EV=%+v", myE2EV)
		// log.Printf("myE2UV=%+v", myE2UV)

		if e1UV.AboutEq(myE1UV) && e2UV.AboutEq(myE2UV) {
			// log.Printf("Found matching face: %v", is.faceInfo.m.dumpFace(baseFaceIdx, is.faces[baseFaceIdx]))
			// Note that the matching face is not the baseFaceIdx! We want the other face on this edge.
			continue
		}

		return baseFaceIdx, true
	}
	return 0, false
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
			log.Printf("Found matching face: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
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
