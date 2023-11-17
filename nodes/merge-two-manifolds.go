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
	log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)
	log.Printf("merge2manifolds: shared edges: %+v", sharedEdges)
	log.Printf("merge2manifolds: shared faces: %+v", sharedFaces)

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
		fi.merge2manisOneEdge(sharedVerts, edge, sharedEdges[edge][0], sharedEdges[edge][1])
	case len(sharedVerts) == 0 && len(sharedEdges) == 0 && len(sharedFaces) == 0: // simple concatenation - no sharing
	default:
		log.Printf("WARNING: merge2manifolds - unhandled shares: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}
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
		log.Printf("WARNING: merge2manisOneEdge: unhandled case: normals don't match: %v vs %v", fi.src.faceNormals[srcFaceIdx], fi.dst.faceNormals[dstFaceIdx])
		return
	}
	if len(fi.src.faces[srcFaceIdx]) != 4 {
		log.Printf("WARNING: merge2manisOneEdge: unhandled case: src.faces[%v] len=%v=%+v", srcFaceIdx, len(fi.src.faces[srcFaceIdx]), fi.src.faces[srcFaceIdx])
		return
	}
	if len(fi.src.faces[srcFaces[1]]) != 4 {
		log.Printf("WARNING: merge2manisOneEdge: unhandled case: src.faces[%v] len=%v=%+v", srcFaces[1], len(fi.src.faces[srcFaces[1]]), fi.src.faces[srcFaces[1]])
		return
	}
	if len(fi.dst.faces[dstFaceIdx]) != 4 {
		log.Printf("WARNING: merge2manisOneEdge: unhandled case: dst.faces[%v] len=%v=%+v", dstFaceIdx, len(fi.dst.faces[dstFaceIdx]), fi.dst.faces[dstFaceIdx])
		return
	}
	if len(fi.dst.faces[dstFaces[1]]) != 4 {
		log.Printf("WARNING: merge2manisOneEdge: unhandled case: dst.faces[%v] len=%v=%+v", dstFaces[1], len(fi.dst.faces[dstFaces[1]]), fi.dst.faces[dstFaces[1]])
		return
	}

	srcFaceToDelete := srcFaces[1]
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
	dstShortEdgeUV := dstShortEdgeVector.Normalized()
	if !srcLongEdgeUV.AboutEq(dstShortEdgeUV) {
		if srcLongEdgeUV.AboutEq(dstShortEdgeUV.Negated()) {
			fi.src.facesTargetedForDeletion[srcFaceToDelete] = true
			fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, edge, nil)
			return
		}

		dstLongEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaces[1])
		log.Printf("merge2manisOneEdge: dstLongEV=%+v", dstLongEV)

		log.Printf("merge2manisOneEdge: srcFaceToDelete=%v, normal=%v", srcFaceToDelete, fi.src.faceNormals[srcFaceToDelete])
		log.Printf("merge2manisOneEdge: dstFaces[1]=%v, normal=%v", dstFaces[1], fi.dst.faceNormals[dstFaces[1]])
		if fi.src.faceNormals[srcFaceToDelete].AboutEq(fi.dst.faceNormals[dstFaces[1]].Negated()) {
			fi.src.facesTargetedForDeletion[srcFaceToDelete] = true
			fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, edge, nil)
			return
		}

		log.Printf("WARNING: merge2manisOneEdge: unhandled case: edge unit vectors don't match: %v vs %v", srcLongEdgeUV, dstShortEdgeUV)
		return
	}
	// log.Printf("merge2manisOneEdge: edge unit vectors match: %v, srcLongEdgeVector=%v, dstShortEdgeVector=%v", srcLongEdgeUV, srcLongEdgeVector, dstShortEdgeVector)

	dstShortConnectedEdge := makeEdge(vertIdx, dstShortOtherVertIdx)

	dstShortNextEV := fi.dst.connectedEdgeVectorFromVertOnFace(dstShortOtherVertIdx, dstShortConnectedEdge, dstFaceIdx)
	dstShortNextVertIdx := dstShortNextEV.toVertIdx
	// log.Printf("merge2manisOneEdge: dstShortNextVertIdx=%v, tmpVec=%v", dstShortNextVertIdx, tmpVec)
	shortenFaceEdge := makeEdge(dstShortOtherVertIdx, dstShortNextVertIdx)
	// log.Printf("merge2manisOneEdge: shortenFaceEdge=%v", shortenFaceEdge)

	fi.src.deleteFaceAndMoveNeighbors(srcFaceToDelete, dstShortEdgeVector)
	fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, shortenFaceEdge, nil)
}

/*
2023/11/16 21:54:01 manifoldMerge: src.badEdges=0=map[]
2023/11/16 21:54:01 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 21:54:01 merge2manifolds: shared verts: map[275:[[0 3 4] [177 178 179]] 276:[[0 2 3] [177 179 180]]]
2023/11/16 21:54:01 merge2manifolds: shared edges: map[[275 276]:[[0 3] [177 179]]]
2023/11/16 21:54:01 merge2manifolds: shared faces: map[]
2023/11/16 21:54:01 faceArea [275 276 301 300]: 4.2354551832614975
2023/11/16 21:54:01 faceArea [304 276 275 305]: 0.9999999999999996
2023/11/16 21:54:01 faceArea [277 276 275 274]: 0.9999999999999994
2023/11/16 21:54:01 faceArea [299 275 276 289]: 2.9999999999999987

2023/11/16 21:54:01 connectedEdgeVectorFromVertOnFace(vertIdx=275, edge=[275 276], faceIdx=0): i=0, pIdx=275, lastVertIdx=300, returning ({3.75 4.50 -6.50}).Sub({0.15 4.50 -7.50})
2023/11/16 21:54:01 merge2manisOneEdge: srcLongEV={edge:[275 300] fromVertIdx:275 toVertIdx:300 toSubFrom:{X:3.59996997898018 Y:0 Z:1.003308720986067} length:3.7371663381719116}

2023/11/16 21:54:01 connectedEdgeVectorFromVertOnFace(vertIdx=275, edge=[275 276], faceIdx=177): i=2, pIdx=275, nextIdx=274, returning ({-0.85 4.50 -7.52}).Sub({0.15 4.50 -7.50})
2023/11/16 21:54:01 merge2manisOneEdge: dstShortEV={edge:[274 275] fromVertIdx:275 toVertIdx:274 toSubFrom:{X:-0.9997998999159141 Y:0 Z:-0.020004002802642695} length:1}

2023/11/16 21:54:01 connectedEdgeVectorFromVertOnFace(vertIdx=275, edge=[275 276], faceIdx=3): i=2, pIdx=275, nextIdx=305, returning ({0.15 5.50 -7.50}).Sub({0.15 4.50 -7.50})
2023/11/16 21:54:01 merge2manisOneEdge: srcShortEV={edge:[275 305] fromVertIdx:275 toVertIdx:305 toSubFrom:{X:0 Y:1 Z:0} length:1}

2023/11/16 22:02:38 connectedEdgeVectorFromVertOnFace(vertIdx=275, edge=[275 276], faceIdx=179): i=1, pIdx=275, lastVertIdx=299, returning ({0.15 7.50 -7.50}).Sub({0.15 4.50 -7.50})
2023/11/16 22:02:38 merge2manisOneEdge: dstLongEV={edge:[275 299] fromVertIdx:275 toVertIdx:299 toSubFrom:{X:0 Y:3 Z:0} length:3}

2023/11/16 22:05:23 merge2manisOneEdge: srcFaceToDelete=3, normal={-1.00 -0.00 -0.02}
2023/11/16 22:05:23 merge2manisOneEdge: dstFaces[1]=179, normal={1.00 0.00 0.02}

2023/11/16 21:54:01 WARNING: merge2manisOneEdge: unhandled case: edge unit vectors don't match: {0.96 0.00 0.27} vs {-1.00 0.00 -0.02}
*/

func (is *infoSetT) deleteFaceAndMoveNeighbors(deleteFaceIdx faceIndexT, move Vec3) {
	log.Printf("BEFORE deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
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
				log.Printf("changing face[%v][%v] from vertIdx=%v to vertIdx=%v", faceIdx, i, vertIdx, newIdx)
				is.faces[faceIdx][i] = newIdx
			}
		}
	}

	is.facesTargetedForDeletion[deleteFaceIdx] = true
	log.Printf("AFTER deleteFaceAndMoveNeighbors(deleteFaceIdx=%v, move=%v), #faces=%v\n%v", deleteFaceIdx, move, len(is.faces), is.faceInfo.m.dumpFaces(is.faces))
}
