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

	log.Printf("merge2manisOneEdge: edge=%v, sorted srcFaces by area desc:\n%v\n%v",
		edge, fi.m.dumpFace(srcFaces[0], fi.src.faces[srcFaces[0]]), fi.m.dumpFace(srcFaces[1], fi.src.faces[srcFaces[1]]))
	log.Printf("merge2manisOneEdge: edge=%v, sorted dstFaces by area asc:\n%v\n%v",
		edge, fi.m.dumpFace(dstFaces[0], fi.dst.faces[dstFaces[0]]), fi.m.dumpFace(dstFaces[1], fi.dst.faces[dstFaces[1]]))
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

		log.Printf("TODO - FIX THIS")
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
	log.Printf("merge2manisOneEdge: edge unit vectors match: %v, srcLongEdgeVector=%v, dstShortEdgeVector=%v", srcLongEdgeUV, srcLongEdgeVector, dstShortEdgeVector)

	dstShortConnectedEdge := makeEdge(vertIdx, dstShortOtherVertIdx)
	log.Printf("merge2manisOneEdge: dstShortConnectedEdge=%v - SUSPECT!!!", dstShortConnectedEdge)

	dstShortNextEV := fi.dst.connectedEdgeVectorFromVertOnFace(dstShortOtherVertIdx, dstShortConnectedEdge, dstFaceIdx)
	dstShortNextVertIdx := dstShortNextEV.toVertIdx
	log.Printf("merge2manisOneEdge: dstShortNextEV=%+v", dstShortNextEV)
	shortenFaceEdge := makeEdge(dstShortOtherVertIdx, dstShortNextVertIdx)
	log.Printf("merge2manisOneEdge: shortenFaceEdge=%v", shortenFaceEdge)

	fi.src.deleteFaceAndMoveNeighbors(srcFaceToDelete, dstShortEdgeVector)
	fi.dst.cutNeighborsAndShortenFaceOnEdge(dstFaceIdx, srcShortEdgeVector, shortenFaceEdge, nil)
}

/*
manifoldMerge: srcFaces=[[0 8 9 3] [0 3 11 10] [0 10 13 8] [8 13 12 9] [9 12 11 3] [10 11 12 13]]
face[0]={[0   8  9  3]}: {{0. 0. 0.} {2. 0. 0.} {2. 0. 1.} {0. 0. 1.}}
face[1]={[0   3 11 10]}: {{0. 0. 0.} {0. 0. 1.} {0. 1. 1.} {0. 1. 0.}}
face[2]={[0  10 13  8]}: {{0. 0. 0.} {0. 1. 0.} {2. 1. 0.} {2. 0. 0.}}
face[3]={[8  13 12  9]}: {{2. 0. 0.} {2. 1. 0.} {2. 1. 1.} {2. 0. 1.}}
face[4]={[9  12 11  3]}: {{2. 0. 1.} {2. 1. 1.} {0. 1. 1.} {0. 0. 1.}}
face[5]={[10 11 12 13]}: {{0. 1. 0.} {0. 1. 1.} {2. 1. 1.} {2. 1. 0.}}

2023/11/17 19:46:51 manifoldMerge: dstFaces=[[0 1 2 3] [0 3 5 4] [0 4 7 1] [1 7 6 2] [2 6 5 3] [4 5 6 7]]
face[0]={[0 1 2 3]}: {{0. 0. 0.} {1. 0. 0.} {1. 0. 1.} {0. 0. 1.}}
face[1]={[0 3 5 4]}: {{0. 0. 0.} {0. 0. 1.} {0. 2. 1.} {0. 2. 0.}}
face[2]={[0 4 7 1]}: {{0. 0. 0.} {0. 2. 0.} {1. 2. 0.} {1. 0. 0.}}
face[3]={[1 7 6 2]}: {{1. 0. 0.} {1. 2. 0.} {1. 2. 1.} {1. 0. 1.}}
face[4]={[2 6 5 3]}: {{1. 0. 1.} {1. 2. 1.} {0. 2. 1.} {0. 0. 1.}}
face[5]={[4 5 6 7]}: {{0. 2. 0.} {0. 2. 1.} {1. 2. 1.} {1. 2. 0.}}

2023/11/17 19:46:51 manifoldMerge: src.badEdges=0=map[]
2023/11/17 19:46:51 manifoldMerge: dst.badEdges=0=map[]

2023/11/17 20:46:54 merge2manisOneEdge: edge=[0 3], sorted srcFaces by area desc:
face[0]={[0 8  9  3]}: {{0. 0. 0.} {2. 0. 0.} {2. 0. 1.} {0. 0. 1.}}
face[1]={[0 3 11 10]}: {{0. 0. 0.} {0. 0. 1.} {0. 1. 1.} {0. 1. 0.}}
2023/11/17 20:46:54 merge2manisOneEdge: edge=[0 3], sorted dstFaces by area asc:
face[0]={[0 1 2 3]}: {{0. 0. 0.} {1. 0. 0.} {1. 0. 1.} {0. 0. 1.}}
face[1]={[0 3 5 4]}: {{0. 0. 0.} {0. 0. 1.} {0. 2. 1.} {0. 2. 0.}}

2023/11/17 19:46:51 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=0): i=0, pIdx=0, nextIdx=8, returning ({2. 0. 0.}).Sub({0. 0. 0.})
2023/11/17 19:46:51 merge2manisOneEdge: srcLongEV={edge:[0 8] fromVertIdx:0 toVertIdx:8 toSubFrom:{X:2 Y:0 Z:0} length:2}
2023/11/17 19:46:51 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=0): i=0, pIdx=0, nextIdx=1, returning ({1. 0. 0.}).Sub({0. 0. 0.})
2023/11/17 19:46:51 merge2manisOneEdge: dstShortEV={edge:[0 1] fromVertIdx:0 toVertIdx:1 toSubFrom:{X:1 Y:0 Z:0} length:1}
2023/11/17 19:46:51 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=1): i=0, pIdx=0, lastVertIdx=10, returning ({0. 1. 0.}).Sub({0. 0. 0.})
2023/11/17 19:46:51 merge2manisOneEdge: srcShortEV={edge:[0 10] fromVertIdx:0 toVertIdx:10 toSubFrom:{X:0 Y:1 Z:0} length:1}

2023/11/17 19:46:51 merge2manisOneEdge: srcLongEdgeUV={1. 0. 0.}
2023/11/17 19:46:51 merge2manisOneEdge: dstShortEdgeUV={1. 0. 0.}

2023/11/17 19:46:51 merge2manisOneEdge: edge unit vectors match: {1. 0. 0.}, srcLongEdgeVector={2. 0. 0.}, dstShortEdgeVector={1. 0. 0.}
2023/11/17 19:46:51 merge2manisOneEdge: dstShortConnectedEdge=[0 1] - SUSPECT!!!

2023/11/17 19:46:51 connectedEdgeVectorFromVertOnFace(vertIdx=1, edge=[0 1], faceIdx=0): i=1, pIdx=1, nextIdx=2, returning ({1. 0. 1.}).Sub({1. 0. 0.})
2023/11/17 19:46:51 merge2manisOneEdge: dstShortNextEV={edge:[1 2] fromVertIdx:1 toVertIdx:2 toSubFrom:{X:0 Y:0 Z:1} length:1}
2023/11/17 19:46:51 merge2manisOneEdge: shortenFaceEdge=[1 2]

2023/11/17 19:46:51 BEFORE deleteFaceAndMoveNeighbors(deleteFaceIdx=1, move={1. 0. 0.}), #faces=6
face[0]={[0 8 9 3]}: {{0. 0. 0.} {2. 0. 0.} {2. 0. 1.} {0. 0. 1.}}
face[1]={[0 3 11 10]}: {{0. 0. 0.} {0. 0. 1.} {0. 1. 1.} {0. 1. 0.}}
face[2]={[0 10 13 8]}: {{0. 0. 0.} {0. 1. 0.} {2. 1. 0.} {2. 0. 0.}}
face[3]={[8 13 12 9]}: {{2. 0. 0.} {2. 1. 0.} {2. 1. 1.} {2. 0. 1.}}
face[4]={[9 12 11 3]}: {{2. 0. 1.} {2. 1. 1.} {0. 1. 1.} {0. 0. 1.}}
face[5]={[10 11 12 13]}: {{0. 1. 0.} {0. 1. 1.} {2. 1. 1.} {2. 1. 0.}}

2023/11/17 19:46:51 changing face[5][0] from vertIdx=10 to vertIdx=15
2023/11/17 19:46:51 changing face[5][1] from vertIdx=11 to vertIdx=14
2023/11/17 19:46:51 changing face[0][0] from vertIdx=0 to vertIdx=1
2023/11/17 19:46:51 changing face[0][3] from vertIdx=3 to vertIdx=2
2023/11/17 19:46:51 changing face[2][0] from vertIdx=0 to vertIdx=1
2023/11/17 19:46:51 changing face[2][1] from vertIdx=10 to vertIdx=15
2023/11/17 19:46:51 changing face[4][2] from vertIdx=11 to vertIdx=14
2023/11/17 19:46:51 changing face[4][3] from vertIdx=3 to vertIdx=2

2023/11/17 19:46:51 AFTER deleteFaceAndMoveNeighbors(deleteFaceIdx=1, move={1. 0. 0.}), #faces=6
face[0]={[1 8 9 2]}: {{1. 0. 0.} {2. 0. 0.} {2. 0. 1.} {1. 0. 1.}}
face[1]={[0 3 11 10]}: {{0. 0. 0.} {0. 0. 1.} {0. 1. 1.} {0. 1. 0.}}
face[2]={[1 15 13 8]}: {{1. 0. 0.} {1. 1. 0.} {2. 1. 0.} {2. 0. 0.}}
face[3]={[8 13 12 9]}: {{2. 0. 0.} {2. 1. 0.} {2. 1. 1.} {2. 0. 1.}}
face[4]={[9 12 14 2]}: {{2. 0. 1.} {2. 1. 1.} {1. 1. 1.} {1. 0. 1.}}
face[5]={[15 14 12 13]}: {{1. 1. 0.} {1. 1. 1.} {2. 1. 1.} {2. 1. 0.}}

2023/11/17 19:46:51 BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=0, move={0. 1. 0.}, edge=[1 2]), #faces=6
face[0]={[0 1 2 3]}: {{0. 0. 0.} {1. 0. 0.} {1. 0. 1.} {0. 0. 1.}}
face[1]={[0 3 5 4]}: {{0. 0. 0.} {0. 0. 1.} {0. 2. 1.} {0. 2. 0.}}
face[2]={[0 4 7 1]}: {{0. 0. 0.} {0. 2. 0.} {1. 2. 0.} {1. 0. 0.}}
face[3]={[1 7 6 2]}: {{1. 0. 0.} {1. 2. 0.} {1. 2. 1.} {1. 0. 1.}}
face[4]={[2 6 5 3]}: {{1. 0. 1.} {1. 2. 1.} {0. 2. 1.} {0. 0. 1.}}
face[5]={[4 5 6 7]}: {{0. 2. 0.} {0. 2. 1.} {1. 2. 1.} {1. 2. 0.}}

2023/11/17 19:46:51 oldVertsToNewMap: map[0:10 1:15 2:14 3:11]
2023/11/17 19:46:51 cutNeighborsAndShortenFaceOnEdge found 4 affected faces: [2 3 4 1]
2023/11/17 19:46:51 changing face[4][0] from vertIdx=2={1. 0. 1.} to vertIdx=14={1. 1. 1.}
2023/11/17 19:46:51 changing face[4][3] from vertIdx=3={0. 0. 1.} to vertIdx=11={0. 1. 1.}
2023/11/17 19:46:51 adding new cut face: [14 11 3 2]
2023/11/17 19:46:51 changing face[1][0] from vertIdx=0={0. 0. 0.} to vertIdx=10={0. 1. 0.}
2023/11/17 19:46:51 changing face[1][1] from vertIdx=3={0. 0. 1.} to vertIdx=11={0. 1. 1.}
2023/11/17 19:46:51 adding new cut face: [0 3 11 10]
2023/11/17 19:46:51 changing face[2][0] from vertIdx=0={0. 0. 0.} to vertIdx=10={0. 1. 0.}
2023/11/17 19:46:51 changing face[2][3] from vertIdx=1={1. 0. 0.} to vertIdx=15={1. 1. 0.}
2023/11/17 19:46:51 adding new cut face: [10 15 1 0]
2023/11/17 19:46:51 changing face[3][0] from vertIdx=1={1. 0. 0.} to vertIdx=15={1. 1. 0.}
2023/11/17 19:46:51 changing face[3][3] from vertIdx=2={1. 0. 1.} to vertIdx=14={1. 1. 1.}
2023/11/17 19:46:51 adding new cut face: [15 14 2 1]

2023/11/17 19:46:51 AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=0, move={0. 1. 0.}, edge=[1 2]), #faces=10
face[0]={[0 1 2 3]}: {{0. 0. 0.} {1. 0. 0.} {1. 0. 1.} {0. 0. 1.}}
face[1]={[10 11 5 4]}: {{0. 1. 0.} {0. 1. 1.} {0. 2. 1.} {0. 2. 0.}}
face[2]={[10 4 7 15]}: {{0. 1. 0.} {0. 2. 0.} {1. 2. 0.} {1. 1. 0.}}
face[3]={[15 7 6 14]}: {{1. 1. 0.} {1. 2. 0.} {1. 2. 1.} {1. 1. 1.}}
face[4]={[14 6 5 11]}: {{1. 1. 1.} {1. 2. 1.} {0. 2. 1.} {0. 1. 1.}}
face[5]={[4 5 6 7]}: {{0. 2. 0.} {0. 2. 1.} {1. 2. 1.} {1. 2. 0.}}
face[6]={[14 11 3 2]}: {{1. 1. 1.} {0. 1. 1.} {0. 0. 1.} {1. 0. 1.}}
face[7]={[0 3 11 10]}: {{0. 0. 0.} {0. 0. 1.} {0. 1. 1.} {0. 1. 0.}}
face[8]={[10 15 1 0]}: {{0. 1. 0.} {1. 1. 0.} {1. 0. 0.} {0. 0. 0.}}
face[9]={[15 14 2 1]}: {{1. 1. 0.} {1. 1. 1.} {1. 0. 1.} {1. 0. 0.}}

2023/11/17 19:46:51 BAD MERGE: before: src badEdges=0
2023/11/17 19:46:51 BAD MERGE: before: dst badEdges=0
2023/11/17 19:46:51 BAD MERGE: after: dst badEdges=4
2023/11/17 19:46:51 NEW BAD EDGE: [14 15]: face[3]={[15 7 6 14]}: {{1. 1. 0.} {1. 2. 0.} {1. 2. 1.} {1. 1. 1.}}
2023/11/17 19:46:51 NEW BAD EDGE: [14 15]: face[9]={[15 14 2 1]}: {{1. 1. 0.} {1. 1. 1.} {1. 0. 1.} {1. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [14 15]: face[14]={[15 14 12 13]}: {{1. 1. 0.} {1. 1. 1.} {2. 1. 1.} {2. 1. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 15]: face[8]={[10 15 1 0]}: {{0. 1. 0.} {1. 1. 0.} {1. 0. 0.} {0. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 15]: face[9]={[15 14 2 1]}: {{1. 1. 0.} {1. 1. 1.} {1. 0. 1.} {1. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 15]: face[11]={[1 15 13 8]}: {{1. 0. 0.} {1. 1. 0.} {2. 1. 0.} {2. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 2]: face[0]={[0 1 2 3]}: {{0. 0. 0.} {1. 0. 0.} {1. 0. 1.} {0. 0. 1.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 2]: face[9]={[15 14 2 1]}: {{1. 1. 0.} {1. 1. 1.} {1. 0. 1.} {1. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [1 2]: face[10]={[1 8 9 2]}: {{1. 0. 0.} {2. 0. 0.} {2. 0. 1.} {1. 0. 1.}}
2023/11/17 19:46:51 NEW BAD EDGE: [2 14]: face[6]={[14 11 3 2]}: {{1. 1. 1.} {0. 1. 1.} {0. 0. 1.} {1. 0. 1.}}
2023/11/17 19:46:51 NEW BAD EDGE: [2 14]: face[9]={[15 14 2 1]}: {{1. 1. 0.} {1. 1. 1.} {1. 0. 1.} {1. 0. 0.}}
2023/11/17 19:46:51 NEW BAD EDGE: [2 14]: face[13]={[9 12 14 2]}: {{2. 0. 1.} {2. 1. 1.} {1. 1. 1.} {1. 0. 1.}}
2023/11/17 19:46:51 Merge: BAD MERGE STOP


OLD:
manifoldMerge: srcFaces=[[0 8 9 3] [0 3 11 10] [0 10 13 8] [8 13 12 9] [9 12 11 3] [10 11 12 13]]
face[0]={[0 8 9 3]}: {{-0.50 -0.50 -0.50} {1.50 -0.50 -0.50} {1.50 -0.50 0.50} {-0.50 -0.50 0.50}}
face[1]={[0 3 11 10]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 0.50 0.50} {-0.50 0.50 -0.50}}
face[2]={[0 10 13 8]}: {{-0.50 -0.50 -0.50} {-0.50 0.50 -0.50} {1.50 0.50 -0.50} {1.50 -0.50 -0.50}}
face[3]={[8 13 12 9]}: {{1.50 -0.50 -0.50} {1.50 0.50 -0.50} {1.50 0.50 0.50} {1.50 -0.50 0.50}}
face[4]={[9 12 11 3]}: {{1.50 -0.50 0.50} {1.50 0.50 0.50} {-0.50 0.50 0.50} {-0.50 -0.50 0.50}}
face[5]={[10 11 12 13]}: {{-0.50 0.50 -0.50} {-0.50 0.50 0.50} {1.50 0.50 0.50} {1.50 0.50 -0.50}}

2023/11/17 18:57:43 manifoldMerge: dstFaces=[[0 1 2 3] [0 3 5 4] [0 4 7 1] [1 7 6 2] [2 6 5 3] [4 5 6 7]]
face[0]={[0 1 2 3]}: {{-0.50 -0.50 -0.50} {0.50 -0.50 -0.50} {0.50 -0.50 0.50} {-0.50 -0.50 0.50}}
face[1]={[0 3 5 4]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 1.50 0.50} {-0.50 1.50 -0.50}}
face[2]={[0 4 7 1]}: {{-0.50 -0.50 -0.50} {-0.50 1.50 -0.50} {0.50 1.50 -0.50} {0.50 -0.50 -0.50}}
face[3]={[1 7 6 2]}: {{0.50 -0.50 -0.50} {0.50 1.50 -0.50} {0.50 1.50 0.50} {0.50 -0.50 0.50}}
face[4]={[2 6 5 3]}: {{0.50 -0.50 0.50} {0.50 1.50 0.50} {-0.50 1.50 0.50} {-0.50 -0.50 0.50}}
face[5]={[4 5 6 7]}: {{-0.50 1.50 -0.50} {-0.50 1.50 0.50} {0.50 1.50 0.50} {0.50 1.50 -0.50}}

2023/11/17 18:57:43 manifoldMerge: src.badEdges=0=map[]
2023/11/17 18:57:43 manifoldMerge: dst.badEdges=0=map[]

2023/11/17 19:24:24 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=0): i=0, pIdx=0, nextIdx=8, returning ({1.50 -0.50 -0.50}).Sub({-0.50 -0.50 -0.50})
2023/11/17 19:24:24 merge2manisOneEdge: srcLongEV={edge:[0 8] fromVertIdx:0 toVertIdx:8 toSubFrom:{X:2 Y:0 Z:0} length:2}
2023/11/17 19:24:24 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=0): i=0, pIdx=0, nextIdx=1, returning ({0.50 -0.50 -0.50}).Sub({-0.50 -0.50 -0.50})
2023/11/17 19:24:24 merge2manisOneEdge: dstShortEV={edge:[0 1] fromVertIdx:0 toVertIdx:1 toSubFrom:{X:1 Y:0 Z:0} length:1}
2023/11/17 19:24:24 connectedEdgeVectorFromVertOnFace(vertIdx=0, edge=[0 3], faceIdx=1): i=0, pIdx=0, lastVertIdx=10, returning ({-0.50 0.50 -0.50}).Sub({-0.50 -0.50 -0.50})
2023/11/17 19:24:24 merge2manisOneEdge: srcShortEV={edge:[0 10] fromVertIdx:0 toVertIdx:10 toSubFrom:{X:0 Y:1 Z:0} length:1}  // THIS IS NOT ALONG THE DST LONG EDGE - interesting

2023/11/17 18:57:43 merge2manisOneEdge: srcLongEdgeUV={1. 0. 0.}
2023/11/17 18:57:43 merge2manisOneEdge: dstShortEdgeUV={1. 0. 0.}
2023/11/17 18:57:43 merge2manisOneEdge: edge unit vectors match: {1. 0. 0.}, srcLongEdgeVector={2. 0. 0.}, dstShortEdgeVector={1. 0. 0.}
2023/11/17 19:24:24 merge2manisOneEdge: dstShortConnectedEdge=[0 1] - SUSPECT!!!
2023/11/17 19:24:24 connectedEdgeVectorFromVertOnFace(vertIdx=1, edge=[0 1], faceIdx=0): i=1, pIdx=1, nextIdx=2, returning ({0.50 -0.50 0.50}).Sub({0.50 -0.50 -0.50})
2023/11/17 18:57:43 merge2manisOneEdge: dstShortNextEV={edge:[1 2] fromVertIdx:1 toVertIdx:2 toSubFrom:{X:0 Y:0 Z:1} length:1}  // WHERE DID THIS COME FROM?
2023/11/17 18:57:43 merge2manisOneEdge: shortenFaceEdge=[1 2]  // WRONG!!! - SHOULD BE [1 15] !!!

2023/11/17 19:09:40 BEFORE deleteFaceAndMoveNeighbors(deleteFaceIdx=1, move={1. 0. 0.}), #faces=6
face[0]={[0 8 9 3]}: {{-0.50 -0.50 -0.50} {1.50 -0.50 -0.50} {1.50 -0.50 0.50} {-0.50 -0.50 0.50}}
face[1]={[0 3 11 10]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 0.50 0.50} {-0.50 0.50 -0.50}}
face[2]={[0 10 13 8]}: {{-0.50 -0.50 -0.50} {-0.50 0.50 -0.50} {1.50 0.50 -0.50} {1.50 -0.50 -0.50}}
face[3]={[8 13 12 9]}: {{1.50 -0.50 -0.50} {1.50 0.50 -0.50} {1.50 0.50 0.50} {1.50 -0.50 0.50}}
face[4]={[9 12 11 3]}: {{1.50 -0.50 0.50} {1.50 0.50 0.50} {-0.50 0.50 0.50} {-0.50 -0.50 0.50}}
face[5]={[10 11 12 13]}: {{-0.50 0.50 -0.50} {-0.50 0.50 0.50} {1.50 0.50 0.50} {1.50 0.50 -0.50}}

2023/11/17 19:09:40 changing face[0][0] from vertIdx=0 to vertIdx=1
2023/11/17 19:09:40 changing face[0][3] from vertIdx=3 to vertIdx=2
2023/11/17 19:09:40 changing face[2][0] from vertIdx=0 to vertIdx=1
2023/11/17 19:09:40 changing face[2][1] from vertIdx=10 to vertIdx=15
2023/11/17 19:09:40 changing face[4][2] from vertIdx=11 to vertIdx=14
2023/11/17 19:09:40 changing face[4][3] from vertIdx=3 to vertIdx=2
2023/11/17 19:09:40 changing face[5][0] from vertIdx=10 to vertIdx=15
2023/11/17 19:09:40 changing face[5][1] from vertIdx=11 to vertIdx=14

2023/11/17 19:09:40 AFTER deleteFaceAndMoveNeighbors(deleteFaceIdx=1, move={1. 0. 0.}), #faces=6
face[0]={[1 8 9 2]}: {{0.50 -0.50 -0.50} {1.50 -0.50 -0.50} {1.50 -0.50 0.50} {0.50 -0.50 0.50}}
face[1]={[0 3 11 10]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 0.50 0.50} {-0.50 0.50 -0.50}}
face[2]={[1 15 13 8]}: {{0.50 -0.50 -0.50} {0.50 0.50 -0.50} {1.50 0.50 -0.50} {1.50 -0.50 -0.50}}
face[3]={[8 13 12 9]}: {{1.50 -0.50 -0.50} {1.50 0.50 -0.50} {1.50 0.50 0.50} {1.50 -0.50 0.50}}
face[4]={[9 12 14 2]}: {{1.50 -0.50 0.50} {1.50 0.50 0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50}}
face[5]={[15 14 12 13]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {1.50 0.50 0.50} {1.50 0.50 -0.50}}

2023/11/17 19:09:40 BEFORE cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=0, move={0. 1. 0.}, edge=[1 2]), #faces=6
face[0]={[0 1 2 3]}: {{-0.50 -0.50 -0.50} {0.50 -0.50 -0.50} {0.50 -0.50 0.50} {-0.50 -0.50 0.50}}
face[1]={[0 3 5 4]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 1.50 0.50} {-0.50 1.50 -0.50}}
face[2]={[0 4 7 1]}: {{-0.50 -0.50 -0.50} {-0.50 1.50 -0.50} {0.50 1.50 -0.50} {0.50 -0.50 -0.50}}
face[3]={[1 7 6 2]}: {{0.50 -0.50 -0.50} {0.50 1.50 -0.50} {0.50 1.50 0.50} {0.50 -0.50 0.50}}
face[4]={[2 6 5 3]}: {{0.50 -0.50 0.50} {0.50 1.50 0.50} {-0.50 1.50 0.50} {-0.50 -0.50 0.50}}
face[5]={[4 5 6 7]}: {{-0.50 1.50 -0.50} {-0.50 1.50 0.50} {0.50 1.50 0.50} {0.50 1.50 -0.50}}

2023/11/17 19:09:40 oldVertsToNewMap: map[0:10 1:15 2:14 3:11]
2023/11/17 19:09:40 cutNeighborsAndShortenFaceOnEdge found 4 affected faces: [1 2 3 4]
2023/11/17 19:09:40 changing face[2][0] from vertIdx=0={-0.50 -0.50 -0.50} to vertIdx=10={-0.50 0.50 -0.50}
2023/11/17 19:09:40 changing face[2][3] from vertIdx=1={0.50 -0.50 -0.50} to vertIdx=15={0.50 0.50 -0.50}
2023/11/17 19:09:40 adding new cut face: [10 15 1 0]
2023/11/17 19:09:40 changing face[3][0] from vertIdx=1={0.50 -0.50 -0.50} to vertIdx=15={0.50 0.50 -0.50}
2023/11/17 19:09:40 changing face[3][3] from vertIdx=2={0.50 -0.50 0.50} to vertIdx=14={0.50 0.50 0.50}
2023/11/17 19:09:40 adding new cut face: [15 14 2 1]
2023/11/17 19:09:40 changing face[4][0] from vertIdx=2={0.50 -0.50 0.50} to vertIdx=14={0.50 0.50 0.50}
2023/11/17 19:09:40 changing face[4][3] from vertIdx=3={-0.50 -0.50 0.50} to vertIdx=11={-0.50 0.50 0.50}
2023/11/17 19:09:40 adding new cut face: [14 11 3 2]
2023/11/17 19:09:40 changing face[1][0] from vertIdx=0={-0.50 -0.50 -0.50} to vertIdx=10={-0.50 0.50 -0.50}
2023/11/17 19:09:40 changing face[1][1] from vertIdx=3={-0.50 -0.50 0.50} to vertIdx=11={-0.50 0.50 0.50}
2023/11/17 19:09:40 adding new cut face: [0 3 11 10]

2023/11/17 19:09:40 AFTER cutNeighborsAndShortenFaceOnEdge(baseFaceIdx=0, move={0. 1. 0.}, edge=[1 2]), #faces=10
face[0]={[0 1 2 3]}: {{-0.50 -0.50 -0.50} {0.50 -0.50 -0.50} {0.50 -0.50 0.50} {-0.50 -0.50 0.50}}
face[1]={[10 11 5 4]}: {{-0.50 0.50 -0.50} {-0.50 0.50 0.50} {-0.50 1.50 0.50} {-0.50 1.50 -0.50}}
face[2]={[10 4 7 15]}: {{-0.50 0.50 -0.50} {-0.50 1.50 -0.50} {0.50 1.50 -0.50} {0.50 0.50 -0.50}}
face[3]={[15 7 6 14]}: {{0.50 0.50 -0.50} {0.50 1.50 -0.50} {0.50 1.50 0.50} {0.50 0.50 0.50}}
face[4]={[14 6 5 11]}: {{0.50 0.50 0.50} {0.50 1.50 0.50} {-0.50 1.50 0.50} {-0.50 0.50 0.50}}
face[5]={[4 5 6 7]}: {{-0.50 1.50 -0.50} {-0.50 1.50 0.50} {0.50 1.50 0.50} {0.50 1.50 -0.50}}
face[6]={[10 15 1 0]}: {{-0.50 0.50 -0.50} {0.50 0.50 -0.50} {0.50 -0.50 -0.50} {-0.50 -0.50 -0.50}}
face[7]={[15 14 2 1]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50} {0.50 -0.50 -0.50}}
face[8]={[14 11 3 2]}: {{0.50 0.50 0.50} {-0.50 0.50 0.50} {-0.50 -0.50 0.50} {0.50 -0.50 0.50}}
face[9]={[0 3 11 10]}: {{-0.50 -0.50 -0.50} {-0.50 -0.50 0.50} {-0.50 0.50 0.50} {-0.50 0.50 -0.50}}

2023/11/17 19:09:40 BAD MERGE: before: src badEdges=0
2023/11/17 19:09:40 BAD MERGE: before: dst badEdges=0
2023/11/17 19:09:40 BAD MERGE: after: dst badEdges=4
2023/11/17 19:09:40 NEW BAD EDGE: [14 15]: face[3]={[15 7 6 14]}: {{0.50 0.50 -0.50} {0.50 1.50 -0.50} {0.50 1.50 0.50} {0.50 0.50 0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [14 15]: face[7]={[15 14 2 1]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50} {0.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [14 15]: face[14]={[15 14 12 13]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {1.50 0.50 0.50} {1.50 0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 15]: face[6]={[10 15 1 0]}: {{-0.50 0.50 -0.50} {0.50 0.50 -0.50} {0.50 -0.50 -0.50} {-0.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 15]: face[7]={[15 14 2 1]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50} {0.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 15]: face[11]={[1 15 13 8]}: {{0.50 -0.50 -0.50} {0.50 0.50 -0.50} {1.50 0.50 -0.50} {1.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [2 14]: face[7]={[15 14 2 1]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50} {0.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [2 14]: face[8]={[14 11 3 2]}: {{0.50 0.50 0.50} {-0.50 0.50 0.50} {-0.50 -0.50 0.50} {0.50 -0.50 0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [2 14]: face[13]={[9 12 14 2]}: {{1.50 -0.50 0.50} {1.50 0.50 0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 2]: face[0]={[0 1 2 3]}: {{-0.50 -0.50 -0.50} {0.50 -0.50 -0.50} {0.50 -0.50 0.50} {-0.50 -0.50 0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 2]: face[7]={[15 14 2 1]}: {{0.50 0.50 -0.50} {0.50 0.50 0.50} {0.50 -0.50 0.50} {0.50 -0.50 -0.50}}
2023/11/17 19:09:40 NEW BAD EDGE: [1 2]: face[10]={[1 8 9 2]}: {{0.50 -0.50 -0.50} {1.50 -0.50 -0.50} {1.50 -0.50 0.50} {0.50 -0.50 0.50}}
2023/11/17 19:09:40 Merge: BAD MERGE STOP
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
