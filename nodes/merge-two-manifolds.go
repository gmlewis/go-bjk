package nodes

import (
	"log"
	"slices"

	"golang.org/x/exp/maps"
)

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	// step 1 - find all shared vertices, edges, and faces
	sharedVerts, sharedEdges, sharedFaces := fi.findSharedVEFs()
	log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)
	log.Printf("merge2manifolds: shared edges: %+v", sharedEdges)
	log.Printf("merge2manifolds: shared faces: %+v", sharedFaces)

	switch {
	case len(sharedFaces) > 0:
		log.Fatalf("merge2manifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedEdges) == 1:
		edges := maps.Keys(sharedEdges)
		edge := edges[0]
		fi.merge2manisOneEdge(sharedVerts, edge, sharedEdges[edge][0], sharedEdges[edge][1])
	case len(sharedVerts) == 0 && len(sharedEdges) == 0 && len(sharedFaces) == 0: // simple concatenation - no sharing
	default:
		log.Fatalf("merge2manifolds - unhandled: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}

	// last step: combine face sets
	fi.m.Faces = append(fi.dst.faces, fi.src.faces...)
}

func (fi *faceInfoT) merge2manisOneEdge(sharedVerts sharedVertsMapT, edge edgeT, srcFaces, dstFaces []faceIndexT) {
	assert(len(srcFaces) == 2 && len(dstFaces) == 2, "merge2manisOneEdge: want 2 srcFaces and 2 dstFaces")

	// sort srcFaces by area (descending - largest first)
	if srcFace0Area, srcFace1Area := fi.m.faceArea(fi.src.faces[srcFaces[0]]), fi.m.faceArea(fi.src.faces[srcFaces[1]]); srcFace0Area < srcFace1Area {
		srcFaces[0], srcFaces[1] = srcFaces[1], srcFaces[0]
	}
	// sort dstFaces by area (ascending - smallest first)
	if dstFace0Area, dstFace1Area := fi.m.faceArea(fi.dst.faces[dstFaces[0]]), fi.m.faceArea(fi.dst.faces[dstFaces[1]]); dstFace0Area > dstFace1Area {
		dstFaces[0], dstFaces[1] = dstFaces[1], dstFaces[0]
	}

	// log.Printf("merge2manisOneEdge: sorted srcFaces by area desc:\n%v\n%v",
	// 	fi.m.dumpFace(srcFaces[0], fi.src.faces[srcFaces[0]]), fi.m.dumpFace(srcFaces[1], fi.src.faces[srcFaces[1]]))
	// log.Printf("merge2manisOneEdge: sorted dstFaces by area asc:\n%v\n%v",
	// 	fi.m.dumpFace(dstFaces[0], fi.dst.faces[dstFaces[0]]), fi.m.dumpFace(dstFaces[1], fi.dst.faces[dstFaces[1]]))
	srcFaceIdx, dstFaceIdx := srcFaces[0], dstFaces[0]

	if !fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx]) {
		log.Fatalf("merge2manisOneEdge: unhandled case: normals don't match: %v vs %v", fi.src.faceNormals[srcFaceIdx], fi.dst.faceNormals[dstFaceIdx])
	}
	if len(fi.src.faces[srcFaceIdx]) != 4 {
		log.Fatalf("merge2manisOneEdge: unhandled case: src.faces[%v] len=%v=%+v", srcFaceIdx, len(fi.src.faces[srcFaceIdx]), fi.src.faces[srcFaceIdx])
	}
	if len(fi.src.faces[srcFaces[1]]) != 4 {
		log.Fatalf("merge2manisOneEdge: unhandled case: src.faces[%v] len=%v=%+v", srcFaces[1], len(fi.src.faces[srcFaces[1]]), fi.src.faces[srcFaces[1]])
	}
	if len(fi.dst.faces[dstFaceIdx]) != 4 {
		log.Fatalf("merge2manisOneEdge: unhandled case: dst.faces[%v] len=%v=%+v", dstFaceIdx, len(fi.dst.faces[dstFaceIdx]), fi.dst.faces[dstFaceIdx])
	}
	if len(fi.dst.faces[dstFaces[1]]) != 4 {
		log.Fatalf("merge2manisOneEdge: unhandled case: dst.faces[%v] len=%v=%+v", dstFaces[1], len(fi.dst.faces[dstFaces[1]]), fi.dst.faces[dstFaces[1]])
	}

	vertIdx := edge[0]
	_, srcLongEdgeVector := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaceIdx)
	dstShortOtherVertIdx, dstShortEdgeVector := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaceIdx)

	// log.Printf("merge2manisOneEdge: srcLongOtherVertIdx=%v, srcLongEdgeVector=%v", srcLongOtherVertIdx, srcLongEdgeVector)
	srcLongEdgeUV := srcLongEdgeVector.Normalized()
	// log.Printf("merge2manisOneEdge: dstShortOtherVertIdx=%v, dstShortEdgeVector=%v", dstShortOtherVertIdx, dstShortEdgeVector)
	dstShortEdgeUV := dstShortEdgeVector.Normalized()
	if !srcLongEdgeUV.AboutEq(dstShortEdgeUV) {
		if srcLongEdgeUV.AboutEq(dstShortEdgeUV.Negated()) {
			log.Fatalf("merge2manisOneEdge: unhandled case: edge unit vectors face opposite each other: %v vs %v", srcLongEdgeUV, dstShortEdgeUV)
		}

		log.Fatalf("merge2manisOneEdge: unhandled case: edge unit vectors don't match: %v vs %v", srcLongEdgeUV, dstShortEdgeUV)
	}

	// log.Printf("merge2manisOneEdge: edge unit vectors match: %v, srcLongEdgeVector=%v, dstShortEdgeVector=%v", srcLongEdgeUV, srcLongEdgeVector, dstShortEdgeVector)

	_, srcShortEdgeVector := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaces[1])
	// log.Printf("merge2manisOneEdge: srcShortOtherVertIdx=%v, srcShortEdgeVector=%v", srcShortOtherVertIdx, srcShortEdgeVector)

	dstShortConnectedEdge := makeEdge(vertIdx, dstShortOtherVertIdx)
	dstShortNextVertIdx, _ := fi.dst.connectedEdgeVectorFromVertOnFace(dstShortOtherVertIdx, dstShortConnectedEdge, dstFaceIdx)
	// log.Printf("merge2manisOneEdge: dstShortNextVertIdx=%v, tmpVec=%v", dstShortNextVertIdx, tmpVec)
	shortenFaceEdge := makeEdge(dstShortOtherVertIdx, dstShortNextVertIdx)
	// log.Printf("merge2manisOneEdge: shortenFaceEdge=%v", shortenFaceEdge)

	fi.src.deleteFaceAndMoveNeighbors(srcFaces[1], dstShortEdgeVector)
	fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, shortenFaceEdge)
}

func (is *infoSetT) deleteFaceAndMoveNeighbors(deleteFaceIdx faceIndexT, move Vec3) {
	log.Printf("BEFORE deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
	face := is.faces[deleteFaceIdx]
	oldVertsToNewMap := is.moveVerts(face, move)
	affectedFaces := map[faceIndexT]bool{}

	for vertIdx := range oldVertsToNewMap {
		for _, faceIdx := range is.vert2Faces[vertIdx] {
			if faceIdx == deleteFaceIdx {
				continue
			}
			affectedFaces[faceIdx] = true
		}
	}

	for faceIdx := range affectedFaces {
		for i, vertIdx := range is.faces[faceIdx] {
			if newIdx, ok := oldVertsToNewMap[vertIdx]; ok {
				log.Printf("changing face[%v][%v] from vertIdx=%v to vertIdx=%v", faceIdx, i, vertIdx, newIdx)
				is.faces[faceIdx][i] = newIdx
			}
		}
	}

	is.faces = slices.Delete(is.faces, int(deleteFaceIdx), int(deleteFaceIdx+1)) // invalidates other faceInfoT maps - last step.
	log.Printf("AFTER deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}

func (is *infoSetT) cutNeighborsAndShortenFaceOnEdge(baseFaceIdx faceIndexT, move Vec3, edge edgeT) {
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
		oldCutFace := make([]VertIndexT, 0, len(face))
		newCutFace := make([]VertIndexT, 0, len(face)/2)
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
		// Fill in the gap (created by moving this face) with a new face.
		slices.Reverse(newCutFace)
		oldCutFace = append(oldCutFace, newCutFace...)
		log.Printf("adding new cut face: %+v", oldCutFace)
		is.faces = append(is.faces, oldCutFace)
	}

	log.Printf("AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=%v, move=%v, edge=%v), #faces=%v\n%v", baseFaceIdx, move, edge, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}

func (is *infoSetT) moveVerts(face FaceT, move Vec3) map[VertIndexT]VertIndexT {
	m := is.faceInfo.m
	uniqueVertsMap := is.uniqueVertsMap()

	vertsOldToNew := make(map[VertIndexT]VertIndexT, len(face))
	for _, vertIdx := range face {
		p := m.Verts[vertIdx].Add(move)
		s := p.String()
		if idx, ok := uniqueVertsMap[s]; ok {
			vertsOldToNew[vertIdx] = idx
			continue
		}
		idx := VertIndexT(len(m.Verts))
		m.Verts = append(m.Verts, p)
		vertsOldToNew[vertIdx] = idx
		uniqueVertsMap[s] = idx
	}

	return vertsOldToNew
}

func (is *infoSetT) uniqueVertsMap() map[string]VertIndexT {
	m := is.faceInfo.m
	uniqueVertsMap := map[string]VertIndexT{}
	for i, vert := range m.Verts {
		s := vert.String()
		uniqueVertsMap[s] = VertIndexT(i)
	}
	log.Printf("uniqueVertsMap found %v unique verts", len(uniqueVertsMap))
	return uniqueVertsMap
}
