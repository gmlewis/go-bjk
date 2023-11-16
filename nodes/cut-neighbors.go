package nodes

import (
	"log"
	"slices"

	"golang.org/x/exp/maps"
)

func (is *infoSetT) cutNeighborsAndShortenFaceOnEdge(baseFaceIdx faceIndexT, move Vec3, edge edgeT, newCutFaceOKToAdd func(FaceT) bool) {
	log.Printf("BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=%v, move=%v, edge=%v), #faces=%v\n%v", baseFaceIdx, move, edge, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
	baseFace := is.faces[baseFaceIdx]
	oldVertsToNewMap := is.moveVerts(baseFace, move)
	affectedFaces := map[faceIndexT]bool{}

	for vertIdx := range oldVertsToNewMap {
		for _, faceIdx := range is.vert2Faces[vertIdx] {
			if faceIdx == baseFaceIdx {
				continue
			}
			affectedFaces[faceIdx] = true
		}
	}
	log.Printf("cutNeighborsAndShortenFaceOnEdge found %v affected faces: %+v", len(affectedFaces), maps.Keys(affectedFaces))

	for faceIdx := range affectedFaces {
		face := is.faces[faceIdx]
		originalFaceNormal := is.faceNormals[faceIdx]
		oldCutFace := make(FaceT, 0, len(face))
		newCutFace := make(FaceT, 0, len(face)/2)
		var faceHasEdge bool
		for i, vertIdx := range face {
			if newIdx, ok := oldVertsToNewMap[vertIdx]; ok {
				log.Printf("changing face[%v][%v] from vertIdx=%v to vertIdx=%v", faceIdx, i, vertIdx, newIdx)
				is.faces[faceIdx][i] = newIdx
				oldCutFace = append(oldCutFace, vertIdx)
				newCutFace = append(newCutFace, newIdx)
			}
			nextIdx := face[(i+1)%len(face)]
			if makeEdge(vertIdx, nextIdx) == edge {
				faceHasEdge = true // leave this face shortened without filling in the gap.
			}
		}
		if faceHasEdge {
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
					log.Fatalf("unable to make new face %+v normal (%v) same as original %+v (%v)", oldCutFace, newFaceNormal, face, originalFaceNormal)
				}
			}

			log.Printf("adding new cut face: %+v", oldCutFace)
			is.faces = append(is.faces, oldCutFace)
		} else {
			log.Printf("NOT ADDING new cut face: %+v !!!", oldCutFace)
		}
	}

	log.Printf("AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=%v, move=%v, edge=%v), #faces=%v\n%v", baseFaceIdx, move, edge, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}

// moveVerts creates new (or reuses old) vertices and returns the mapping from the
// old face's vertIndexes to the new vertices, without modifying the face.
func (is *infoSetT) moveVerts(face FaceT, move Vec3) map[VertIndexT]VertIndexT {
	m := is.faceInfo.m

	vertsOldToNew := make(map[VertIndexT]VertIndexT, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx].Add(move)
		newVertIdx := m.AddVert(v)
		vertsOldToNew[vertIdx] = newVertIdx
	}

	return vertsOldToNew
}
