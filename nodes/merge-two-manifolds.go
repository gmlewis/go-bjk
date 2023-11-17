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
	log.Printf("merge2manisOneEdge: srcLongEdgeUV=%v", srcLongEdgeUV)
	dstShortEdgeUV := dstShortEdgeVector.Normalized()
	log.Printf("merge2manisOneEdge: dstShortEdgeUV=%v", dstShortEdgeUV)

	if !fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx]) {
		if !fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx].Negated()) {
			log.Printf("WARNING: merge2manisOneEdge: unhandled case: normals don't match: %v vs %v", fi.src.faceNormals[srcFaceIdx], fi.dst.faceNormals[dstFaceIdx])
			return
		}

		// TODO - FIX THIS
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
2023/11/16 23:08:31 manifoldMerge: src.badEdges=0=map[]
2023/11/16 23:08:31 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 23:08:31 merge2manifolds: shared verts: map[422:[[0 1 2] [295 296 299]] 423:[[0 1 4] [295 296 297]]]
2023/11/16 23:08:31 merge2manifolds: shared edges: map[[422 423]:[[0 1] [295 296]]]
2023/11/16 23:08:31 merge2manifolds: shared faces: map[]
2023/11/16 23:08:31 faceArea [431 430 422 423]: 7.9043000032484105
2023/11/16 23:08:31 faceArea [432 423 422 433]: 97.17639415758327
2023/11/16 23:08:31 faceArea [425 424 423 422]: 9.764135298130396
2023/11/16 23:08:31 faceArea [426 422 423 427]: 8.83421765068939
2023/11/16 23:08:31 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=1): i=2, pIdx=422, nextIdx=433, returning ({9.49 16.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 23:08:31 merge2manisOneEdge: srcLongEV={edge:[422 433] fromVertIdx:422 toVertIdx:433 toSubFrom:{X:0 Y:11 Z:0} length:11}
2023/11/16 23:08:31 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=296): i=1, pIdx=422, lastVertIdx=426, returning ({9.49 6.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 23:08:31 merge2manisOneEdge: dstShortEV={edge:[422 426] fromVertIdx:422 toVertIdx:426 toSubFrom:{X:0 Y:1 Z:0} length:1}
2023/11/16 23:08:31 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=0): i=2, pIdx=422, lastVertIdx=430, returning ({8.49 5.50 0.34}).Sub({9.49 5.50 0.38})
2023/11/16 23:08:31 merge2manisOneEdge: srcShortEV={edge:[422 430] fromVertIdx:422 toVertIdx:430 toSubFrom:{X:-0.9991996797437448 Y:0 Z:-0.04000000000000037} length:1.000000000000001}
2023/11/16 23:08:31 merge2manisOneEdge: srcLongEdgeUV={0.00 1.00 0.00}
2023/11/16 23:08:31 merge2manisOneEdge: dstShortEdgeUV={0.00 1.00 0.00}
2023/11/16 23:08:31 BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=1, move={0.00 1.00 0.00}, edge=[422 423]), #faces=6  // CORRECT EDGE - CORRECT MOVE - WRONG baseFaceIdx - should be 0

face[0]={[431 430 422 423]}: {{4.54 5.50 7.19} {8.49 5.50 0.34} {9.49 5.50 0.38} {5.08 5.50 8.03}}
face[1]={[432 423 422 433]}: {{5.08 16.50 8.03} {5.08 5.50 8.03} {9.49 5.50 0.38} {9.49 16.50 0.38}}  // TARGET face
face[2]={[433 422 430 434]}: {{9.49 16.50 0.38} {9.49 5.50 0.38} {8.49 5.50 0.34} {8.49 16.50 0.34}}
face[3]={[434 430 431 435]}: {{8.49 16.50 0.34} {8.49 5.50 0.34} {4.54 5.50 7.19} {4.54 16.50 7.19}}
face[4]={[435 431 423 432]}: {{4.54 16.50 7.19} {4.54 5.50 7.19} {5.08 5.50 8.03} {5.08 16.50 8.03}}
face[5]={[432 433 434 435]}: {{5.08 16.50 8.03} {9.49 16.50 0.38} {8.49 16.50 0.34} {4.54 16.50 7.19}}

2023/11/16 23:08:31 cutNeighborsAndShortenFaceOnEdge found 4 affected faces: [2 5 4 0]   // WRONG!!!  Should be [0 2 3 4]
2023/11/16 23:08:31 changing face[0][2] from vertIdx=422={9.49 5.50 0.38} to vertIdx=426={9.49 6.50 0.38}
2023/11/16 23:08:31 changing face[0][3] from vertIdx=423={5.08 5.50 8.03} to vertIdx=427={5.08 6.50 8.03}
2023/11/16 23:08:31 changing face[2][0] from vertIdx=433={9.49 16.50 0.38} to vertIdx=437={9.49 17.50 0.38}  // WRONG!
2023/11/16 23:08:31 changing face[2][1] from vertIdx=422={9.49 5.50 0.38} to vertIdx=426={9.49 6.50 0.38}
2023/11/16 23:08:31 WARNING: unable to make new face [437 426 422 433] normal ({0.00 0.00 -0.00}) same as original [437 426 430 434] ({0.04 0.00 -1.00}), skipping
2023/11/16 23:08:31 changing face[5][0] from vertIdx=432={5.08 16.50 8.03} to vertIdx=436={5.08 17.50 8.03}  // WRONG!
2023/11/16 23:08:31 changing face[5][1] from vertIdx=433={9.49 16.50 0.38} to vertIdx=437={9.49 17.50 0.38}  // WRONG!
2023/11/16 23:08:31 WARNING: unable to make new face [436 437 433 432] normal ({-0.87 0.00 -0.50}) same as original [436 437 434 435] ({-0.00 1.00 0.00}), skipping
2023/11/16 23:08:31 changing face[4][2] from vertIdx=423={5.08 5.50 8.03} to vertIdx=427={5.08 6.50 8.03}
2023/11/16 23:08:31 changing face[4][3] from vertIdx=432={5.08 16.50 8.03} to vertIdx=436={5.08 17.50 8.03}  // WRONG!
2023/11/16 23:08:31 WARNING: unable to make new face [427 436 432 423] normal ({0.00 0.00 0.00}) same as original [435 431 427 436] ({-0.85 0.00 0.53}), skipping
2023/11/16 23:08:31 AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=1, move={0.00 1.00 0.00}, edge=[422 423]), #faces=6
face[0]={[431 430 426 427]}: {{4.54 5.50 7.19} {8.49 5.50 0.34} {9.49 6.50 0.38} {5.08 6.50 8.03}}
face[1]={[432 423 422 433]}: {{5.08 16.50 8.03} {5.08 5.50 8.03} {9.49 5.50 0.38} {9.49 16.50 0.38}}
face[2]={[437 426 430 434]}: {{9.49 17.50 0.38} {9.49 6.50 0.38} {8.49 5.50 0.34} {8.49 16.50 0.34}}
face[3]={[434 430 431 435]}: {{8.49 16.50 0.34} {8.49 5.50 0.34} {4.54 5.50 7.19} {4.54 16.50 7.19}}
face[4]={[435 431 427 436]}: {{4.54 16.50 7.19} {4.54 5.50 7.19} {5.08 6.50 8.03} {5.08 17.50 8.03}}
face[5]={[436 437 434 435]}: {{5.08 17.50 8.03} {9.49 17.50 0.38} {8.49 16.50 0.34} {4.54 16.50 7.19}}
2023/11/16 23:08:31

DELETING FACE!!! face[296]={[426 422 423 427]}: {{9.49 6.50 0.38} {9.49 5.50 0.38} {5.08 5.50 8.03} {5.08 6.50 8.03}}
2023/11/16 23:08:31
*/

/*
2023/11/16 23:00:06 manifoldMerge: src.badEdges=0=map[]
2023/11/16 23:00:06 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 23:00:06 merge2manifolds: shared verts: map[422:[[0 1 2] [295 296 299]] 423:[[0 1 4] [295 296 297]]]
2023/11/16 23:00:06 merge2manifolds: shared edges: map[[422 423]:[[0 1] [295 296]]]
2023/11/16 23:00:06 merge2manifolds: shared faces: map[]
2023/11/16 23:00:06 faceArea [431 430 422 423]: 7.9043000032484105
2023/11/16 23:00:06 faceArea [432 423 422 433]: 97.17639415758327
2023/11/16 23:00:06 faceArea [425 424 423 422]: 9.764135298130396
2023/11/16 23:00:06 faceArea [426 422 423 427]: 8.83421765068939
2023/11/16 23:00:06 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=1): i=2, pIdx=422, nextIdx=433, returning ({9.49 16.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 23:00:06 merge2manisOneEdge: srcLongEV={edge:[422 433] fromVertIdx:422 toVertIdx:433 toSubFrom:{X:0 Y:11 Z:0} length:11}
2023/11/16 23:00:06 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=296): i=1, pIdx=422, lastVertIdx=426, returning ({9.49 6.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 23:00:06 merge2manisOneEdge: dstShortEV={edge:[422 426] fromVertIdx:422 toVertIdx:426 toSubFrom:{X:0 Y:1 Z:0} length:1}
2023/11/16 23:00:06 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=0): i=2, pIdx=422, lastVertIdx=430, returning ({8.49 5.50 0.34}).Sub({9.49 5.50 0.38})
2023/11/16 23:00:06 merge2manisOneEdge: srcShortEV={edge:[422 430] fromVertIdx:422 toVertIdx:430 toSubFrom:{X:-0.9991996797437448 Y:0 Z:-0.04000000000000037} length:1.000000000000001}
2023/11/16 23:00:06 merge2manisOneEdge: srcLongEdgeUV={0.00 1.00 0.00}
2023/11/16 23:00:06 merge2manisOneEdge: dstShortEdgeUV={0.00 1.00 0.00}

2023/11/16 23:00:06 BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=1, move={0.00 1.00 0.00}, edge=[422 423]), #faces=6

face[0]={[431 430 422 423]}: {{4.54 5.50 7.19} {8.49 5.50 0.34} {9.49 5.50 0.38} {5.08 5.50 8.03}}
face[1]={[432 423 422 433]}: {{5.08 16.50 8.03} {5.08 5.50 8.03} {9.49 5.50 0.38} {9.49 16.50 0.38}}
face[2]={[433 422 430 434]}: {{9.49 16.50 0.38} {9.49 5.50 0.38} {8.49 5.50 0.34} {8.49 16.50 0.34}}
face[3]={[434 430 431 435]}: {{8.49 16.50 0.34} {8.49 5.50 0.34} {4.54 5.50 7.19} {4.54 16.50 7.19}}
face[4]={[435 431 423 432]}: {{4.54 16.50 7.19} {4.54 5.50 7.19} {5.08 5.50 8.03} {5.08 16.50 8.03}}
face[5]={[432 433 434 435]}: {{5.08 16.50 8.03} {9.49 16.50 0.38} {8.49 16.50 0.34} {4.54 16.50 7.19}}

2023/11/16 23:00:06 cutNeighborsAndShortenFaceOnEdge found 4 affected faces: [5 4 0 2]

2023/11/16 23:00:06 changing face[0][2] from vertIdx=422 to vertIdx=426
2023/11/16 23:00:06 changing face[0][3] from vertIdx=423 to vertIdx=427
2023/11/16 23:00:06 changing face[2][0] from vertIdx=433 to vertIdx=437
2023/11/16 23:00:06 changing face[2][1] from vertIdx=422 to vertIdx=426

2023/11/16 23:00:06 unable to make new face [437 426 422 433] normal ({0.00 0.00 -0.00}) same as original [437 426 430 434] ({0.04 0.00 -1.00})
*/

/*
2023/11/16 22:44:00 manifoldMerge: src.badEdges=0=map[]
2023/11/16 22:44:00 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 22:44:00 merge2manifolds: shared verts: map[422:[[0 1 2] [295 296 299]] 423:[[0 1 4] [295 296 297]]]
2023/11/16 22:44:00 merge2manifolds: shared edges: map[[422 423]:[[0 1] [295 296]]]
2023/11/16 22:44:00 merge2manifolds: shared faces: map[]
2023/11/16 22:44:00 faceArea [431 430 422 423]: 7.9043000032484105
2023/11/16 22:44:00 faceArea [432 423 422 433]: 97.17639415758327
2023/11/16 22:44:00 faceArea [425 424 423 422]: 9.764135298130396
2023/11/16 22:44:00 faceArea [426 422 423 427]: 8.83421765068939

2023/11/16 22:44:00 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=1): i=2, pIdx=422, nextIdx=433, returning ({9.49 16.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 22:44:00 merge2manisOneEdge: srcLongEV={edge:[422 433] fromVertIdx:422 toVertIdx:433 toSubFrom:{X:0 Y:11 Z:0} length:11}

2023/11/16 22:44:00 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=296): i=1, pIdx=422, lastVertIdx=426, returning ({9.49 6.50 0.38}).Sub({9.49 5.50 0.38})
2023/11/16 22:44:00 merge2manisOneEdge: dstShortEV={edge:[422 426] fromVertIdx:422 toVertIdx:426 toSubFrom:{X:0 Y:1 Z:0} length:1}

2023/11/16 22:44:00 connectedEdgeVectorFromVertOnFace(vertIdx=422, edge=[422 423], faceIdx=0): i=2, pIdx=422, lastVertIdx=430, returning ({8.49 5.50 0.34}).Sub({9.49 5.50 0.38})
2023/11/16 22:44:00 merge2manisOneEdge: srcShortEV={edge:[422 430] fromVertIdx:422 toVertIdx:430 toSubFrom:{X:-0.9991996797437448 Y:0 Z:-0.04000000000000037} length:1.000000000000001}

2023/11/16 22:44:00 merge2manisOneEdge: srcLongEdgeUV={0.00 1.00 0.00}
2023/11/16 22:44:00 merge2manisOneEdge: dstShortEdgeUV={0.00 1.00 0.00}
*/

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
