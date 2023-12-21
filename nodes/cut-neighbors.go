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

func (fi *faceInfoT) cutAndShortenEdge(edge edgeT, srcFaceIndices, dstFaceIndices []faceIndexT) {
	if len(srcFaceIndices) != 2 || len(dstFaceIndices) != 2 {
		log.Printf("WARNING: cutAndShortenEdge: unhandled case len(srcFaceIndices)=%v != 2 || len(dstFaceIndices)=%v != 2", len(srcFaceIndices), len(dstFaceIndices))
		return
	}

	// log.Printf("\n\ncutAndShortenEdge(edge=%v, srcFaceIndices=%+v, dstFaceIndices=%+v)", edge, srcFaceIndices, dstFaceIndices)
	srcFaceIdx0, srcFaceIdx1 := srcFaceIndices[0], srcFaceIndices[1]
	// log.Printf("cutAndShortenEdge: srcFaceIdx0 normal: %v, %v", fi.src.faceNormals[srcFaceIdx0], fi.m.dumpFace(srcFaceIdx0, fi.src.faces[srcFaceIdx0]))
	// log.Printf("cutAndShortenEdge: srcFaceIdx1 normal: %v, %v", fi.src.faceNormals[srcFaceIdx1], fi.m.dumpFace(srcFaceIdx1, fi.src.faces[srcFaceIdx1]))
	dstFaceIdx0, dstFaceIdx1 := dstFaceIndices[0], dstFaceIndices[1]
	// log.Printf("cutAndShortenEdge: dstFaceIdx0 normal: %v, %v", fi.dst.faceNormals[dstFaceIdx0], fi.m.dumpFace(dstFaceIdx0, fi.dst.faces[dstFaceIdx0]))
	// log.Printf("cutAndShortenEdge: dstFaceIdx1 normal: %v, %v", fi.dst.faceNormals[dstFaceIdx1], fi.m.dumpFace(dstFaceIdx1, fi.dst.faces[dstFaceIdx1]))

	negSrcFaceNormal0 := fi.src.faceNormals[srcFaceIdx0].Negated()
	negSrcFaceNormal1 := fi.src.faceNormals[srcFaceIdx1].Negated()
	dstFaceNormal0 := fi.dst.faceNormals[dstFaceIdx0]
	dstFaceNormal1 := fi.dst.faceNormals[dstFaceIdx1]

	srcFaceArea0 := fi.m.faceArea(fi.src.faces[srcFaceIdx0])
	srcFaceArea1 := fi.m.faceArea(fi.src.faces[srcFaceIdx1])
	dstFaceArea0 := fi.m.faceArea(fi.dst.faces[dstFaceIdx0])
	dstFaceArea1 := fi.m.faceArea(fi.dst.faces[dstFaceIdx1])

	// sharedEdges := sharedEdgesMapT{edge: [2][]faceIndexT{{srcFaceIdx0, srcFaceIdx1}, {dstFaceIdx0, dstFaceIdx1}}}

	switch {
	case negSrcFaceNormal0.AboutEq(dstFaceNormal0):
		log.Printf("WARNING: NOT IMPLEMENTED YET: cutAndShortenEdge case00: srcFaceArea0=%0.2f, dstFaceArea0=%0.2f", srcFaceArea0, dstFaceArea0)
		// fi.removeCompleteOverlapOnEdge(edge, srcFaceIdx0, dstFaceIdx0)
	case negSrcFaceNormal0.AboutEq(dstFaceNormal1):
		log.Printf("WARNING: NOT IMPLEMENTED YET: cutAndShortenEdge case01: srcFaceArea0=%0.2f, dstFaceArea1=%0.2f", srcFaceArea0, dstFaceArea1)
		// fi.removeCompleteOverlapOnEdge(edge, srcFaceIdx0, dstFaceIdx1)
	case negSrcFaceNormal1.AboutEq(dstFaceNormal0):
		log.Printf("WARNING: NOT IMPLEMENTED YET: cutAndShortenEdge case10: srcFaceArea1=%0.2f, dstFaceArea0=%0.2f", srcFaceArea1, dstFaceArea0)
		// fi.removeCompleteOverlapOnEdge(edge, srcFaceIdx1, dstFaceIdx0)
	case negSrcFaceNormal1.AboutEq(dstFaceNormal1):
		// log.Printf("cutAndShortenEdge case11: srcFaceIdx=%v area=%0.2f, dstFaceIdx=%v area=%0.2f", srcFaceIdx1, srcFaceArea1, dstFaceIdx1, dstFaceArea1)
		if srcFaceArea1 < dstFaceArea1 { // delete srcFaceIdx1 and cut dstFaceArea1
			fi.src.facesTargetedForDeletion[srcFaceIdx1] = true
			moveDstEVs := fi.src.makeEdgeVectors(edge, srcFaceIdx1)
			// log.Printf("moveDstEVs[0]=%+v", moveDstEVs[0])
			// log.Printf("moveDstEVs[1]=%+v", moveDstEVs[1])
			affectedDstEVs := fi.dst.makeEdgeVectors(edge, dstFaceIdx1)
			// log.Printf("affectedDstEVs[0]=%+v", affectedDstEVs[0])
			// log.Printf("affectedDstEVs[1]=%+v", affectedDstEVs[1])

			affectedEdges := make([]edgeT, 0, 2)
			switch {
			case !AboutEq(affectedDstEVs[0].length, moveDstEVs[0].length):
				affectedEdges = append(affectedEdges, affectedDstEVs[0].edge)
			case !AboutEq(affectedDstEVs[1].length, moveDstEVs[0].length):
				affectedEdges = append(affectedEdges, affectedDstEVs[1].edge)
			}

			fi.dst.resizeFace(nil, dstFaceIdx1, affectedEdges, moveDstEVs)
		} else { // delete dstFaceArea1 and cut srcFaceArea1
			log.Printf("WARNING: NOT IMPLEMENTED YET: cutAndShortenEdge case11 B: srcFaceArea1=%0.2f, dstFaceArea1=%0.2f", srcFaceArea1, dstFaceArea1)
		}
	default:
		log.Printf("WARNING: cutAndShortenEdge: unhandled case - no equal normals")
	}
}
