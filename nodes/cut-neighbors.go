// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"
)

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
