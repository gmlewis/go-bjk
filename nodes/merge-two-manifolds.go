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
	case len(sharedFaces) > 0:
		log.Fatalf("merge2manifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedEdges) == 1:
		edges := maps.Keys(sharedEdges)
		edge := edges[0]
		fi.merge2manisOneEdge(sharedVerts, edge, sharedEdges[edge][0], sharedEdges[edge][1])
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

	log.Printf("merge2manisOneEdge: sorted srcFaces by area desc:\n%v\n%v",
		fi.m.dumpFace(srcFaces[0], fi.src.faces[srcFaces[0]]), fi.m.dumpFace(srcFaces[1], fi.src.faces[srcFaces[1]]))
	log.Printf("merge2manisOneEdge: sorted dstFaces by area asc:\n%v\n%v",
		fi.m.dumpFace(dstFaces[0], fi.dst.faces[dstFaces[0]]), fi.m.dumpFace(dstFaces[1], fi.dst.faces[dstFaces[1]]))
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
	srcLongOtherVertIdx, srcLongEdgeVector := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaceIdx)
	dstShortOtherVertIdx, dstShortEdgeVector := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaceIdx)

	log.Printf("merge2manisOneEdge: srcLongOtherVertIdx=%v, srcLongEdgeVector=%v", srcLongOtherVertIdx, srcLongEdgeVector)
	srcLongEdgeUV := srcLongEdgeVector.Normalized()
	log.Printf("merge2manisOneEdge: dstShortOtherVertIdx=%v, dstShortEdgeVector=%v", dstShortOtherVertIdx, dstShortEdgeVector)
	dstShortEdgeUV := dstShortEdgeVector.Normalized()
	if !srcLongEdgeUV.AboutEq(dstShortEdgeUV) {
		log.Fatalf("merge2manisOneEdge: unhandled case: edge unit vectors don't match: %v vs %v", srcLongEdgeUV, dstShortEdgeUV)
	}

	log.Printf("merge2manisOneEdge: edge unit vectors match: %v, srcLongEdgeVector=%v, dstShortEdgeVector=%v", srcLongEdgeUV, srcLongEdgeVector, dstShortEdgeVector)

	srcShortOtherVertIdx, srcShortEdgeVector := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcFaces[1])
	log.Printf("merge2manisOneEdge: srcShortOtherVertIdx=%v, srcShortEdgeVector=%v", srcShortOtherVertIdx, srcShortEdgeVector)
	dstLongOtherVertIdx, dstLongEdgeVector := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstFaces[1])
	log.Printf("merge2manisOneEdge: dstLongOtherVertIdx=%v, dstLongEdgeVector=%v", dstLongOtherVertIdx, dstLongEdgeVector)

	// fi.src.deleteFaceAndMoveNeighbors(srcFaces[1], dstShortEdgeVector)
	// fi.dst.cutNeighborsAndShortenFace(dstFaceIdx, dstShortEdgeVector)
}
