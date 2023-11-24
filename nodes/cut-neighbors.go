package nodes

import (
	"log"
	"slices"
)

func (is *infoSetT) cutNeighborsAndShortenAlongEdges(baseFaceIdx faceIndexT, e1EV, e2EV edgeVectorT) {
	edge := makeEdge(e1EV.fromVertIdx, e2EV.fromVertIdx)
	amount := 0.5 * (e1EV.length + e2EV.length)
	log.Printf("cutNeighborsAndShortenAlongEdges: baseFaceIdx=%v, e1EV=%v, e2EV=%v", baseFaceIdx, e1EV, e2EV)
	log.Printf("cutNeighborsAndShortenAlongEdges: amount=%v, edge=%v, face %v",
		amount, edge, is.faceInfo.m.dumpFace(baseFaceIdx, is.faces[baseFaceIdx]))

	oldVertsToNewMap, shortenedFaces := is.moveVertsAlongEdgeLoop(baseFaceIdx, amount)
	log.Printf("cutNeighborsAndShortenAlongEdges: oldVertsToNewMap: %+v", oldVertsToNewMap)

	// shortenedFaces := faceSetT{}
	// for vertIdx := range oldVertsToNewMap {
	// 	for _, faceIdx := range is.vertToFaces[vertIdx] {
	// 		if faceIdx == baseFaceIdx {
	// 			continue
	// 		}
	//
	// 		// log.Printf("cutNeighborsAndShortenAlongEdges: affectedFace: vertIdx=%v, %v", vertIdx, is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
	// 		shortenedFaces[faceIdx] = struct{}{}
	// 	}
	// }

	for faceIdx := range shortenedFaces {
		log.Printf("cutNeighborsAndShortenAlongEdges: cutting face %v", faceIdx)
		is.moveFaceVertsAndAddFaceUnlessOnEdge(faceIdx, oldVertsToNewMap, edge)
	}
}

func (is *infoSetT) moveFaceVertsAndAddFaceUnlessOnEdge(faceIdx faceIndexT, oldVertsToNewMap map[VertIndexT]VertIndexT, avoidEdge edgeT) {
	face := is.faces[faceIdx]
	cutFace := append(FaceT{}, face...)
	faceNormal := is.faceNormals[faceIdx]
	log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: face: %v", is.faceInfo.m.dumpFace(faceIdx, face))
	log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: avoidEdge=%v, faceNormal=%v", avoidEdge, faceNormal)

	oldCutFace := make(FaceT, 0, len(face))
	newCutFace := make(FaceT, 0, len(face)/2)

	var skipAddingFace bool
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		faceEdge := makeEdge(vertIdx, nextIdx)
		log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: faceEdge=%v", faceEdge)
		if faceEdge == avoidEdge {
			log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: SKIP ADDING FACE due to face on avoidEdge")
			skipAddingFace = true
		}
		if moveIdx, ok := oldVertsToNewMap[vertIdx]; ok {
			log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: moving old vertIdx=%v to new moveIdx=%v", vertIdx, moveIdx)
			cutFace[i] = moveIdx
			oldCutFace = append(oldCutFace, vertIdx)
			newCutFace = append(newCutFace, moveIdx)
		}
	}

	is.faces[faceIdx] = cutFace
	log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: shortened face: %v", is.faceInfo.m.dumpFace(faceIdx, cutFace))
	log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: oldCutFace: %v", is.faceInfo.m.dumpFace(faceIdx, oldCutFace))
	log.Printf("moveFaceEdgeAndAddFaceUnlessOnEdge: newCutFace: %v", is.faceInfo.m.dumpFace(faceIdx, newCutFace))

	if !skipAddingFace {
		slices.Reverse(newCutFace)
		newFace := append(oldCutFace, newCutFace...)
		newFaceNormal := is.faceInfo.m.CalcFaceNormal(newFace)
		if !newFaceNormal.AboutEq(faceNormal) {
			log.Printf("reversing new face because newFaceNormal=%v and was expecting faceNormal=%v", newFaceNormal, faceNormal)
			slices.Reverse(newFace)
			newFaceNormal = is.faceInfo.m.CalcFaceNormal(newFace)
			if !newFaceNormal.AboutEq(faceNormal) {
				log.Fatalf("moveFaceEdgeAndAddFaceUnlessOnEdge: programming error: new face normal=%v, expecting %v", newFaceNormal, faceNormal)
			}
		}
		log.Printf("Adding new face: normal=%v, %v", newFaceNormal, is.faceInfo.m.dumpFace(faceIndexT(len(is.faces)), newFace))
		is.faces = append(is.faces, newFace)
	}
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
