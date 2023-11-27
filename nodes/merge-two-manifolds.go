// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"

	"golang.org/x/exp/maps"
)

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	// step 1 - find all shared vertices, edges, and faces
	sharedVerts, sharedEdges, sharedFaces := fi.findSharedVEFs()
	// log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)
	// log.Printf("merge2manifolds: shared edges: %+v", sharedEdges)
	// log.Printf("merge2manifolds: shared faces: %+v", sharedFaces)

	switch {
	case len(sharedFaces) > 1:
		log.Printf("WARNING: merge2manifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedFaces) == 1:
		key := maps.Keys(sharedFaces)[0]
		fi.merge2manisOneFace(sharedEdges, sharedFaces[key][0], sharedFaces[key][1])
	case len(sharedEdges) > 1:
		fi.merge2manisManyEdges(sharedEdges)
	case len(sharedEdges) == 1:
		edges := maps.Keys(sharedEdges)
		edge := edges[0]
		fi.merge2manisOneEdge(edge, sharedEdges[edge][0], sharedEdges[edge][1])
	case len(sharedVerts) == 0 && len(sharedEdges) == 0 && len(sharedFaces) == 0: // simple concatenation - no sharing
	default:
		log.Printf("WARNING: merge2manifolds - unhandled shares: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}
}

func (fi *faceInfoT) merge2manisOneEdge(edge edgeT, srcFaces, dstFaces []faceIndexT) {
	assert(len(srcFaces) == 2 && len(dstFaces) == 2, "merge2manisOneEdge: want 2 srcFaces and 2 dstFaces")

	if len(fi.src.faces[srcFaces[0]]) != 4 || len(fi.src.faces[srcFaces[1]]) != 4 {
		// swap src and dst so that src is the quad.
		fi.swapSrcAndDst(nil)
		srcFaces, dstFaces = dstFaces, srcFaces
	}

	// sort srcFaces by area (descending - largest first)
	if srcFace0Area, srcFace1Area := fi.m.faceArea(fi.src.faces[srcFaces[0]]), fi.m.faceArea(fi.src.faces[srcFaces[1]]); srcFace0Area < srcFace1Area {
		srcFaces[0], srcFaces[1] = srcFaces[1], srcFaces[0]
	}
	// sort dstFaces by area (ascending - smallest first)
	if dstFace0Area, dstFace1Area := fi.m.faceArea(fi.dst.faces[dstFaces[0]]), fi.m.faceArea(fi.dst.faces[dstFaces[1]]); dstFace0Area > dstFace1Area {
		dstFaces[0], dstFaces[1] = dstFaces[1], dstFaces[0]
	}

	log.Printf("merge2manisOneEdge: edge=%v, sorted srcFaces by area desc:\n%v\n%v",
		edge, fi.m.dumpFace(srcFaces[0], fi.src.faces[srcFaces[0]]), fi.m.dumpFace(srcFaces[1], fi.src.faces[srcFaces[1]]))
	log.Printf("merge2manisOneEdge: edge=%v, sorted dstFaces by area asc:\n%v\n%v",
		edge, fi.m.dumpFace(dstFaces[0], fi.dst.faces[dstFaces[0]]), fi.m.dumpFace(dstFaces[1], fi.dst.faces[dstFaces[1]]))
	srcFaceIdx, dstFaceIdx := srcFaces[0], dstFaces[0]
	log.Printf("merge2manisOneEdge: edge=%v, srcFaces[0]=srcFaceIdx=%v normal: %v", edge, srcFaceIdx, fi.src.faceNormals[srcFaces[0]])
	log.Printf("merge2manisOneEdge: edge=%v, srcFaces[1]=%v normal: %v", edge, srcFaces[1], fi.src.faceNormals[srcFaces[1]])
	log.Printf("merge2manisOneEdge: edge=%v, dstFaces[0]=dstFaceIdx=%v normal: %v", edge, dstFaceIdx, fi.dst.faceNormals[dstFaces[0]])
	log.Printf("merge2manisOneEdge: edge=%v, dstFaces[1]=%v normal: %v", edge, dstFaces[1], fi.dst.faceNormals[dstFaces[1]])

	// if len(fi.dst.faces[dstFaceIdx]) != 4 || len(fi.dst.faces[dstFaces[1]]) != 4 {
	// 	fi.merge2manisManyEdges(sharedEdges)
	// 	return
	// }

	srcFaceToDelete := srcFaces[1]
	log.Printf("merge2manisOneEdge: srcFaceToDelete=%v", srcFaceToDelete)
	vertIdx := edge[0]
	srcLongEV := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaceIdx)
	srcLongEdgeVector := srcLongEV.toSubFrom
	log.Printf("merge2manisOneEdge: srcLongEV=%+v", srcLongEV)

	dstShortEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaceIdx)
	dstShortOtherVertIdx, dstShortEdgeVector := dstShortEV.toVertIdx, dstShortEV.toSubFrom
	log.Printf("merge2manisOneEdge: dstShortEV=%+v", dstShortEV)

	srcShortEV := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaceToDelete)
	srcShortEdgeVector := srcShortEV.toSubFrom
	log.Printf("merge2manisOneEdge: srcShortEV=%+v", srcShortEV)

	srcLongEdgeUV := srcLongEdgeVector.Normalized()
	log.Printf("merge2manisOneEdge: srcLongEdgeUV=%v", srcLongEdgeUV)
	dstShortEdgeUV := dstShortEdgeVector.Normalized()
	log.Printf("merge2manisOneEdge: dstShortEdgeUV=%v", dstShortEdgeUV)

	if !fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx]) {
		if !fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx].Negated()) {

			// check other normals
			if fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaces[1]].Negated()) {
				log.Printf("merge2manisOneEdge: src face %v abuts to dst face %v", srcFaceIdx, dstFaces[1])

				dstLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaces[1])
				log.Printf("merge2manisOneEdge: dstLongEV=%+v", dstLongEV)

				if srcLongEV.length < dstLongEV.length {
					fi.src.facesTargetedForDeletion[srcFaceIdx] = true

					srcOtherLongEV := fi.src.connectedEdgeVectorFromVertOnFace(edge[1], edge, srcFaceIdx)
					log.Printf("merge2manisOneEdge: srcOtherLongEV=%+v", srcOtherLongEV)

					fi.dst.cutNeighborsAndShortenAlongEdges(dstFaceIdx, srcLongEV, srcOtherLongEV)
					// fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcLongEdgeVector, edge, nil)
					return
				}

				if srcLongEV.length > dstLongEV.length {
					fi.dst.facesTargetedForDeletion[dstFaces[1]] = true

					dstOtherLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(edge[1], edge, dstFaces[1])
					log.Printf("merge2manisOneEdge: dstOtherLongEV=%+v", dstOtherLongEV)

					fi.src.cutNeighborsAndShortenAlongEdges(srcFaces[1], dstLongEV, dstOtherLongEV)
					// fi.src.cutNeighborsAndShortenFaceOnEdge(srcFaces[1], dstLongEV.toSubFrom, edge, nil)
					return
				}
			}

			if fi.src.faceNormals[srcFaces[1]].AboutEq(fi.dst.faceNormals[dstFaceIdx].Negated()) {
				log.Printf("merge2manisOneEdge: src face %v abuts to dst face %v", srcFaces[1], dstFaceIdx)

				dstLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaceIdx)
				log.Printf("merge2manisOneEdge: dstLongEV=%+v", dstLongEV)

				if srcShortEV.length < dstLongEV.length {
					fi.src.facesTargetedForDeletion[srcFaces[1]] = true

					srcOtherShortEV := fi.src.connectedEdgeVectorFromVertOnFace(edge[1], edge, srcFaces[1])
					log.Printf("merge2manisOneEdge: srcOtherShortEV=%+v", srcOtherShortEV)

					fi.dst.cutNeighborsAndShortenAlongEdges(dstFaces[1], srcShortEV, srcOtherShortEV)
					return
				}

				if srcShortEV.length > dstLongEV.length {
					fi.dst.facesTargetedForDeletion[dstFaceIdx] = true

					dstOtherLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(edge[1], edge, dstFaces[1])
					log.Printf("merge2manisOneEdge: dstOtherLongEV=%+v", dstOtherLongEV)

					fi.src.cutNeighborsAndShortenAlongEdges(srcFaceIdx, dstLongEV, dstOtherLongEV)
					// fi.src.cutNeighborsAndShortenFaceOnEdge(srcFaceIdx, dstLongEV.toSubFrom, edge, nil)
					return
				}
			}

			log.Printf("merge2manisOneEdge: face normal srcFaceIdx=%v: %v", srcFaceIdx, fi.src.faceNormals[srcFaceIdx])
			log.Printf("merge2manisOneEdge: face normal srcFaces[1]=%v: %v", srcFaces[1], fi.src.faceNormals[srcFaces[1]])
			log.Printf("merge2manisOneEdge: face normal dstFaceIdx=%v: %v", dstFaceIdx, fi.dst.faceNormals[dstFaceIdx])
			log.Printf("merge2manisOneEdge: face normal dstFaces[1]=%v: %v", dstFaces[1], fi.dst.faceNormals[dstFaces[1]])

			log.Printf("WARNING: merge2manisOneEdge: unhandled case: normals don't match: %v vs %v", fi.src.faceNormals[srcFaceIdx], fi.dst.faceNormals[dstFaceIdx])
			return
		}

		fi.src.cutNeighborsAndShortenFaceOnEdge(srcFaceToDelete, dstShortEdgeVector, edge, nil)
		fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
		return
	}

	if !srcLongEdgeUV.AboutEq(dstShortEdgeUV) {
		if srcLongEdgeUV.AboutEq(dstShortEdgeUV.Negated()) {
			fi.src.facesTargetedForDeletion[srcFaceToDelete] = true
			fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, edge, nil)
			return
		}

		// dstLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaces[1])
		// log.Printf("merge2manisOneEdge: dstLongEV=%+v", dstLongEV)

		// log.Printf("merge2manisOneEdge: srcFaceToDelete=%v, normal=%v", srcFaceToDelete, fi.src.faceNormals[srcFaceToDelete])
		// log.Printf("merge2manisOneEdge: dstFaces[1]=%v, normal=%v", dstFaces[1], fi.dst.faceNormals[dstFaces[1]])
		if fi.src.faceNormals[srcFaceToDelete].AboutEq(fi.dst.faceNormals[dstFaces[1]].Negated()) {
			fi.src.facesTargetedForDeletion[srcFaceToDelete] = true

			// fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, edge, nil)  // old method

			srcE1UV := srcShortEV.toSubFrom.Normalized()
			// log.Printf("merge2manisOneEdge: srcE1UV=%+v", srcE1UV)
			srcE2EV := fi.src.connectedEdgeVectorFromVertOnFace(edge[1], edge, srcFaceToDelete)
			// log.Printf("merge2manisOneEdge: srcE2EV=%+v", srcE2EV)
			srcE2UV := srcE2EV.toSubFrom.Normalized()
			// log.Printf("merge2manisOneEdge: srcE2UV=%+v", srcE2UV)

			baseDstFaceIdx, ok := fi.dst.findBaseFaceSharingTwoEdgeUVs(edge, srcE1UV, srcE2UV)
			if !ok {
				log.Printf("WARNING: merge2manisOneEdge: unhandled case: could not find dst base on edge %v with UVs: %v, %v", edge, srcE1UV, srcE2UV)
				return
			}
			// log.Printf("merge2manisOneEdge: baseDstFaceIdx=%v", baseDstFaceIdx)
			fi.dst.cutNeighborsAndShortenAlongEdges(baseDstFaceIdx, srcShortEV, srcE2EV)
			return
		}

		log.Printf("WARNING: merge2manisOneEdge: unhandled case: edge unit vectors don't match: %v vs %v", srcLongEdgeUV, dstShortEdgeUV)
		return
	}
	// log.Printf("merge2manisOneEdge: edge unit vectors match: %v, srcLongEdgeVector=%v, dstShortEdgeVector=%v", srcLongEdgeUV, srcLongEdgeVector, dstShortEdgeVector)

	dstShortConnectedEdge := makeEdge(vertIdx, dstShortOtherVertIdx)
	// 	log.Printf("merge2manisOneEdge: dstShortConnectedEdge=%v", dstShortConnectedEdge)

	dstShortNextEV := fi.dst.connectedEdgeVectorFromVertOnFace(dstShortOtherVertIdx, dstShortConnectedEdge, dstFaceIdx)
	dstShortNextVertIdx := dstShortNextEV.toVertIdx
	// log.Printf("merge2manisOneEdge: dstShortNextEV=%+v", dstShortNextEV)
	shortenFaceEdge := makeEdge(dstShortOtherVertIdx, dstShortNextVertIdx)
	// log.Printf("merge2manisOneEdge: shortenFaceEdge=%v", shortenFaceEdge)

	fi.src.deleteFaceAndMoveNeighbors(srcFaceToDelete, dstShortEdgeVector)
	fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, shortenFaceEdge, nil)
}

func (is *infoSetT) deleteFaceAndMoveNeighbors(deleteFaceIdx faceIndexT, move Vec3) {
	// log.Printf("BEFORE deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
	face := is.faces[deleteFaceIdx]
	oldVertsToNewMap := is.moveVerts(face, move)
	affectedFaces := map[faceIndexT]bool{}

	for vertIdx := range oldVertsToNewMap {
		for _, faceIdx := range is.vertToFaces[vertIdx] {
			if faceIdx == deleteFaceIdx {
				continue
			}
			affectedFaces[faceIdx] = true
		}
	}

	for faceIdx := range affectedFaces {
		for i, vertIdx := range is.faces[faceIdx] {
			if newIdx, ok := oldVertsToNewMap[vertIdx]; ok {
				// log.Printf("changing face[%v][%v] from vertIdx=%v to vertIdx=%v", faceIdx, i, vertIdx, newIdx)
				is.faces[faceIdx][i] = newIdx
			}
		}
	}

	is.facesTargetedForDeletion[deleteFaceIdx] = true
	// log.Printf("AFTER deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}
