// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"
)

func (is *infoSetT) moveVertsAlongEdgeLoop(faceIdx faceIndexT, amount float64) (vToVMap, faceSetT) {
	log.Fatalf("TEMPORARILY RETURNING NIL: moveVertsAlongEdgeLoop")
	return nil, nil
}

func (is *infoSetT) cutNeighborsAndShortenAlongEdges(baseFaceIdx faceIndexT, e1EV, e2EV edgeVectorT) {
	if e1EV.fromVertIdx == e2EV.fromVertIdx {
		is.cutNeighborsAndShortenAlongTwoConnectedEdges(e1EV, e2EV)
		return
	}

	edge := makeEdge(e1EV.fromVertIdx, e2EV.fromVertIdx)
	amount := 0.5 * (e1EV.length + e2EV.length)
	log.Printf("\n\ncutNeighborsAndShortenAlongEdges: baseFaceIdx=%v, e1EV=%v, e2EV=%v", baseFaceIdx, e1EV, e2EV)
	log.Printf("cutNeighborsAndShortenAlongEdges: amount=%v, edge=%v, face %v",
		amount, edge, is.faceInfo.m.dumpFace(baseFaceIdx, is.faces[baseFaceIdx]))

	// oldVertsToNewMap, shortenedFaces := is.moveVertsAlongEdgeLoop(baseFaceIdx, amount)
	// log.Printf("cutNeighborsAndShortenAlongEdges: oldVertsToNewMap: %+v", oldVertsToNewMap)
	// log.Printf("cutNeighborsAndShortenAlongEdges: %v faces to shorten", len(shortenedFaces))
	//
	// var i int
	// for faceIdx := range shortenedFaces {
	// 	i++
	// 	log.Printf("\n\ncutNeighborsAndShortenAlongEdges: cutting face #%v of %v: faceIdx=%v", i, len(shortenedFaces), faceIdx)
	// 	is.moveFaceVertsAndAddFaceUnlessOnEdge(faceIdx, oldVertsToNewMap, edge)
	// }
	log.Fatalf("TEMPORARILY STOPPING: cutNeighborsAndShortenAlongEdges")
}

func (is *infoSetT) moveFaceVertsAndAddFaceUnlessOnEdge(faceIdx faceIndexT, oldVertsToNewMap map[VertIndexT]VertIndexT, avoidEdge edgeT) {
	is.moveFaceVertsAndAddFace(faceIdx, oldVertsToNewMap, &avoidEdge)
}

func (is *infoSetT) moveFaceVertsAndAddFace(faceIdx faceIndexT, oldVertsToNewMap map[VertIndexT]VertIndexT, avoidEdge *edgeT) {
	face := is.faces[faceIdx]
	cutFace := append(FaceT{}, face...)
	faceNormal := is.faceNormals[faceIdx]
	log.Printf("moveFaceEdgeAndAddFace: face: %v", is.faceInfo.m.dumpFace(faceIdx, face))
	log.Printf("moveFaceEdgeAndAddFace: avoidEdge=%v", avoidEdge)
	log.Printf("moveFaceEdgeAndAddFace: faceNormal=%v", faceNormal)

	oldCutFace := make(FaceT, 0, len(face))
	newCutFace := make(FaceT, 0, len(face)/2)

	var skipAddingFace bool
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		faceEdge := makeEdge(vertIdx, nextIdx)
		log.Printf("moveFaceEdgeAndAddFace: faceEdge=%v", faceEdge)
		if avoidEdge != nil && faceEdge == *avoidEdge {
			log.Printf("moveFaceEdgeAndAddFace: SKIPPING ADDING NEW FACE due to face on avoidEdge")
			skipAddingFace = true
		}
		if moveIdx, ok := oldVertsToNewMap[vertIdx]; ok {
			log.Printf("moveFaceEdgeAndAddFace: moving old vertIdx=%v %v of face %v to new moveIdx=%v %v", vertIdx, is.faceInfo.m.Verts[vertIdx], faceIdx, moveIdx, is.faceInfo.m.Verts[moveIdx])
			cutFace[i] = moveIdx
			if !faceIsValid(cutFace) {
				log.Printf("WARNING: moveFaceEdgeAndAddFace: SKIPPING creation of degenerate face and marking for deletion: %v", is.faceInfo.m.dumpFace(faceIdx, cutFace))
				is.facesTargetedForDeletion[faceIdx] = true
				return
			}
			oldCutFace = append(oldCutFace, vertIdx)
			newCutFace = append(newCutFace, moveIdx)
		}
	}

	is.faces[faceIdx] = cutFace
	log.Printf("moveFaceEdgeAndAddFace: shortened face: %v", is.faceInfo.m.dumpFace(faceIdx, cutFace))
	log.Printf("moveFaceEdgeAndAddFace: oldCutFace: %v", is.faceInfo.m.dumpFace(faceIdx, oldCutFace))
	log.Printf("moveFaceEdgeAndAddFace: newCutFace: %v", is.faceInfo.m.dumpFace(faceIdx, newCutFace))

	if !skipAddingFace {
		slices.Reverse(newCutFace)
		newFace := append(oldCutFace, newCutFace...)
		if len(newFace) < 4 {
			log.Printf("WARNING: moveFaceEdgeAndAddFace: unable to make new face (want >=4 verts): %v", is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
			return
		}
		newFaceNormal := is.faceInfo.m.CalcFaceNormal(newFace)
		log.Printf("moveFaceEdgeAndAddFace: newFace: %v", is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
		log.Printf("moveFaceEdgeAndAddFace: newFace normal: %v", newFaceNormal)
		if !newFaceNormal.AboutEq(faceNormal) {
			log.Printf("reversing new face because newFaceNormal=%v and was expecting faceNormal=%v", newFaceNormal, faceNormal)
			slices.Reverse(newFace)
			newFaceNormal = is.faceInfo.m.CalcFaceNormal(newFace)
			log.Printf("moveFaceEdgeAndAddFace: take 2: newFace: %v", is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
			log.Printf("moveFaceEdgeAndAddFace: take 2: newFace normal: %v", newFaceNormal)
			if !newFaceNormal.AboutEq(faceNormal) {
				log.Printf("WARNING! moveFaceEdgeAndAddFace: programming error: new face normal=%v, expecting %v", newFaceNormal, faceNormal)
				return
			}
		}

		if faceIsValid(newFace) {
			log.Printf("Adding new face: normal=%v, %v\n\n", newFaceNormal, is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
			is.faces = append(is.faces, newFace)
		} else {
			log.Printf("WARNING: moveFaceVertsAndAddFace: attempted to create (SKIPPED) invalid face: normal=%v, %v", newFaceNormal, is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
		}
	}
}

func faceIsValid(face FaceT) bool {
	verts := make(map[VertIndexT]struct{}, len(face))
	for _, vertIdx := range face {
		if _, ok := verts[vertIdx]; ok {
			return false
		}
		verts[vertIdx] = struct{}{}
	}
	return true
}

func (is *infoSetT) cutNeighborsAndShortenFaceOnEdge(baseFaceIdx faceIndexT, move Vec3, edge edgeT, newCutFaceOKToAdd func(FaceT) bool) {
	// log.Printf("BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=%v, move=%v, edge=%v), #faces=%v\n%v", baseFaceIdx, move, edge, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
	baseFace := is.faces[baseFaceIdx]
	oldVertsToNewMap := is.moveVerts(baseFace, move)
	// log.Printf("oldVertsToNewMap: %+v", oldVertsToNewMap)
	affectedFaces := map[faceIndexT]bool{}

	for vertIdx := range oldVertsToNewMap {
		for _, faceIdx := range is.vertToFaces[vertIdx] {
			if faceIdx == baseFaceIdx {
				continue
			}
			affectedFaces[faceIdx] = true
		}
	}
	// log.Printf("cutNeighborsAndShortenFaceOnEdge found %v affected faces: %+v", len(affectedFaces), maps.Keys(affectedFaces))
	// verts := is.faceInfo.m.Verts

	for faceIdx := range affectedFaces {
		face := is.faces[faceIdx]
		originalFaceNormal := is.faceNormals[faceIdx]
		oldCutFace := make(FaceT, 0, len(face))
		newCutFace := make(FaceT, 0, len(face)/2)
		for i, vertIdx := range face {
			if newIdx, ok := oldVertsToNewMap[vertIdx]; ok {
				// log.Printf("changing face[%v][%v] from vertIdx=%v=%v to vertIdx=%v=%v", faceIdx, i, vertIdx, verts[vertIdx], newIdx, verts[newIdx])
				is.faces[faceIdx][i] = newIdx
				oldCutFace = append(oldCutFace, vertIdx)
				newCutFace = append(newCutFace, newIdx)
			}
		}
		// log.Printf("cutNeighborsAndShortenFaceOnEdge: edge=%v, oldCutFace=%+v, newCutFace=%+v", edge, oldCutFace, newCutFace)
		if (len(oldCutFace) >= 2 && makeEdge(oldCutFace[0], oldCutFace[1]) == edge) ||
			(len(newCutFace) >= 2 && makeEdge(newCutFace[0], newCutFace[1]) == edge) {
			// log.Printf("cutNeighborsAndShortenFaceOnEdge: will NOT create a face on this edge! oldCutFace=%v, newCutFace=%v", oldCutFace, newCutFace)
			continue
		}

		if newCutFaceOKToAdd == nil || newCutFaceOKToAdd(oldCutFace) {
			// Fill in the gap (created by moving this face) with a new face.
			// NOTE that this new face MUST face the same direction (have the same normal) as its shortened face above!!!
			slices.Reverse(newCutFace)
			oldCutFace = append(oldCutFace, newCutFace...)
			newFaceNormal := is.faceInfo.m.CalcFaceNormal(oldCutFace)
			if !newFaceNormal.AboutEq(originalFaceNormal) {
				slices.Reverse(oldCutFace)
				newFaceNormal = is.faceInfo.m.CalcFaceNormal(oldCutFace)
				if !newFaceNormal.AboutEq(originalFaceNormal) {
					log.Printf("WARNING: unable to make new face %+v normal (%v) same as original %+v (%v), skipping", oldCutFace, newFaceNormal, face, originalFaceNormal)
					continue
				}
			}

			// log.Printf("adding new cut face: %+v", oldCutFace)
			is.faces = append(is.faces, oldCutFace)
			// } else {
			// 	log.Printf("NOT ADDING new cut face: %+v !!!", oldCutFace)
		}
	}

	// log.Printf("AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=%v, move=%v, edge=%v), #faces=%v\n%v", baseFaceIdx, move, edge, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}

func (is *infoSetT) cutNeighborsAndShortenAlongTwoConnectedEdges(e1EV, e2EV edgeVectorT) {
	if e1EV.fromVertIdx != e2EV.fromVertIdx {
		log.Fatalf("cutNeighborsAndShortenAlongTwoConnectedEdges: programming error: expect two edges to share the same fromVertIdx: (e1EV=%v, e2EV=%v)", e1EV.edge, e2EV.edge)
	}

	// First, find the two base faces, and cut the one with the fewest edges first.
	baseFaceE1Idx := is.findBaseFaceForEdgeVector(e1EV)
	baseFaceE2Idx := is.findBaseFaceForEdgeVector(e2EV)
	if len(is.faces[baseFaceE1Idx]) > len(is.faces[baseFaceE2Idx]) {
		baseFaceE1Idx, baseFaceE2Idx = baseFaceE2Idx, baseFaceE1Idx
		e1EV, e2EV = e2EV, e1EV
	}
	// 	log.Printf("cutNeighborsAndShortenAlongTwoConnectedEdges: baseFaceE1Idx=%v (%v verts), baseFaceE2Idx=%v (%v verts)",
	// 		baseFaceE1Idx, len(is.faces[baseFaceE1Idx]), baseFaceE2Idx, len(is.faces[baseFaceE2Idx]))

	avoidEdge := e1EV.edge
	if baseFaceE1Idx != baseFaceE2Idx {
		amount := e1EV.length
		// log.Printf("cnasatce: amount=%v, e1EV=%v", amount, e1EV)
		// 	m := is.faceInfo.m
		// log.Printf("cnasatce: first: found base face: %v", m.dumpFace(baseFaceE1Idx, is.faces[baseFaceE1Idx]))

		oldVertsToNewMap, shortenedFaces := is.moveVertsAlongEdgeLoop(baseFaceE1Idx, amount)
		// 	log.Printf("cutNeighborsAndShortenAlongEdges: first: oldVertsToNewMap: %+v", oldVertsToNewMap)

		for faceIdx := range shortenedFaces {
			// log.Printf("cutNeighborsAndShortenAlongEdges: first: cutting face %v", m.dumpFace(faceIdx, is.faces[faceIdx]))
			is.moveFaceVertsAndAddFace(faceIdx, oldVertsToNewMap, nil)
		}

		// Since the above operation modified all the map lookups, we need to update this infoset.
		// But we also need to preserve all the targeted faces for deletion.
		// log.Printf("cnasatce: regenerating infoSetT after first base face cut")
		newIS := is.faceInfo.genFaceInfoForSet(is.faces)
		if is == is.faceInfo.dst {
			newIS.facesTargetedForDeletion = is.faceInfo.dst.facesTargetedForDeletion
			is.faceInfo.dst = newIS
		} else {
			newIS.facesTargetedForDeletion = is.faceInfo.src.facesTargetedForDeletion
			is.faceInfo.src = newIS
		}
		is = newIS
		// And we also need to recalculate baseFaceE2Idx.
		baseFaceE2Idx = is.findBaseFaceForEdgeVector(e2EV)
		if baseFaceE2Idx == baseFaceE1Idx {
			log.Printf("WARNING: Already cut using base face %v", baseFaceE2Idx)
		}
	} else {
		avoidEdge = makeEdge(e1EV.fromVertIdx, e2EV.fromVertIdx)
	}

	// Now cut along the other edge.
	amount := e2EV.length
	// log.Printf("cnasatce: second: found base face: %v", m.dumpFace(baseFaceE2Idx, is.faces[baseFaceE2Idx]))
	// Identify the edge on this base face that should be avoided when creating a new face.
	//	log.Printf("cnasatce: second: badEdges=%+v", is.badEdges)
	//	log.Printf("cnasatce: second: badFaces=%+v", is.badFaces)

	oldVertsToNewMap, shortenedFaces := is.moveVertsAlongEdgeLoop(baseFaceE2Idx, amount)
	//	log.Printf("cutNeighborsAndShortenAlongEdges: second: oldVertsToNewMap: %+v", oldVertsToNewMap)

	for faceIdx := range shortenedFaces {
		//		log.Printf("cutNeighborsAndShortenAlongEdges: second: cutting face %v", m.dumpFace(faceIdx, is.faces[faceIdx]))
		is.moveFaceVertsAndAddFaceUnlessOnEdge(faceIdx, oldVertsToNewMap, avoidEdge)
	}
}

func (is *infoSetT) findBaseFaceForEdgeVector(refEV edgeVectorT) faceIndexT {
	// log.Printf("\n\nfindBaseFaceForEdgeVector: refEV=%v", refEV)
	// m := is.faceInfo.m
nextFace:
	for _, faceIdx := range is.vertToFaces[refEV.fromVertIdx] {
		// log.Printf("fbffev: looking at face %v", m.dumpFace(faceIdx, is.faces[faceIdx]))
		face := is.faces[faceIdx]
		for i, vIdx := range face {
			nextIdx := face[(i+1)%len(face)]
			if vIdx != refEV.fromVertIdx && nextIdx != refEV.fromVertIdx {
				continue
			}

			otherVertIdx := vIdx
			if otherVertIdx == refEV.fromVertIdx {
				otherVertIdx = nextIdx
			}

			ev := is.faceInfo.m.makeEdgeVector(refEV.fromVertIdx, otherVertIdx)
			// log.Printf("fbffev: looking at ev %v", ev)

			dotProduct := Vec3Dot(refEV.toSubFrom, ev.toSubFrom)
			// log.Printf("fbffev: dotProduct=%0.2f", dotProduct)
			if dotProduct < 1 { // check next edge
				continue
			}

			if dotProduct >= 1 {
				// log.Printf("fbffev: eliminating faceIdx=%v as a possible base face", faceIdx)
				continue nextFace
			}
		}
		return faceIdx
	}
	log.Fatalf("findBaseFaceForEdgeVector: programming error")
	return 0
}
