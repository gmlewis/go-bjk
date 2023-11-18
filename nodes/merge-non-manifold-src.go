package nodes

import (
	"log"
)

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh (ideally, although it is possible that it is still non-manifold).
func (fi *faceInfoT) mergeNonManifoldSrc() {
	// If there are N bad edges and N bad faces, chances are good that these are simply open
	// (unconnected) extrusions looking to join the dst mesh.
	if len(fi.src.badEdges) == len(fi.src.badFaces) {
		fi.connectOpenSrcExtrusionsToDst()
		return
	}

	// srcFaceIndicesToEdges := reverseMapBadEdges(fi.src.badEdges)
	// debugFaces := make([]FaceT, 0, len(srcFaceIndicesToEdges))
	// log.Printf("mergeNonManifoldSrc: srcFaceIndicesToEdges: %+v", srcFaceIndicesToEdges)
	// for srcFaceIdx, badEdges := range srcFaceIndicesToEdges {
	// 	debugFaces = append(debugFaces, fi.src.faces[srcFaceIdx])
	// 	log.Printf("mergeNonManifoldSrc: src.faces[%v] has %v bad edges: %+v", srcFaceIdx, len(badEdges), badEdges)
	// }
	// fi.m.Faces = debugFaces
	// fi.m.WriteSTL(fmt.Sprintf("debug-%v-%v-badFaces-%v.stl", len(fi.src.faces), len(fi.dst.faces), len(debugFaces)))

	for edge, faceIdxes := range fi.src.badEdges {
		switch len(faceIdxes) {
		case 1:
			log.Printf("WARNING: mergeNonManifoldSrc: skipping edge %v with one face", edge)
		case 3:
			// debugFileBaseName := fmt.Sprintf("debug-%v-%v-edge-%v-%v-", len(fi.src.faces), len(fi.dst.faces), edge[0], edge[1])
			// log.Printf("debugFileBaseName=%v", debugFileBaseName)
			// fi.m.Faces = fi.src.faces
			// fi.m.WriteSTL(debugFileBaseName + "src.stl")
			// fi.m.Faces = fi.dst.faces
			// fi.m.WriteSTL(debugFileBaseName + "dst.stl")

			fi.src.fixEdge3Faces(edge, faceIdxes)
		default:
			log.Printf("WARNING: mergeNonManifoldSrc: skipping edge %v with %v faces", edge, len(faceIdxes))
		}
	}
}

func (is *infoSetT) fixEdge3Faces(edge edgeT, faceIdxes []faceIndexT) {
	f0, f1, f2 := faceIdxes[0], faceIdxes[1], faceIdxes[2]
	switch {
	case is.faceNormals[f0].AboutEq(is.faceNormals[f1]):
		is.fixEdge2OverlapingFaces(edge, f0, f1, f2)
	case is.faceNormals[f1].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f1, f2, f0)
	case is.faceNormals[f0].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f0, f2, f1)
	default:
		log.Printf("WARNING: fixEdge3Faces: unhandled case normals: %v %v %v", is.faceNormals[f0], is.faceNormals[f1], is.faceNormals[f2])
	}
}

func (is *infoSetT) fixEdge2OverlapingFaces(edge edgeT, f0, f1, otherFaceIdx faceIndexT) {
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f0=%v: %v", edge, f0, is.faceNormals[f0], is.faceInfo.m.dumpFace(f0, is.faces[f0]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f1=%v: %v", edge, f1, is.faceNormals[f1], is.faceInfo.m.dumpFace(f1, is.faces[f1]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, otherFace=%v: %v", edge, otherFaceIdx, is.faceNormals[otherFaceIdx], is.faceInfo.m.dumpFace(otherFaceIdx, is.faces[otherFaceIdx]))
	// log.Printf("fixEdge2OverlapingFaces(edge=%v): shared edge: %v - %v", edge, is.faceInfo.m.Verts[edge[0]], is.faceInfo.m.Verts[edge[1]])
	//
	// is.faceInfo.m.Faces = []FaceT{is.faces[f0]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "f0.stl")
	// is.faceInfo.m.Faces = []FaceT{is.faces[f1]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "f1.stl")
	// is.faceInfo.m.Faces = []FaceT{is.faces[otherFaceIdx]}
	// is.faceInfo.m.WriteSTL(debugFileBaseName + "otherFaceIdx.stl")

	log.Fatalf("fixEdge2OverlapingFaces: STOP")

	// is.facesTargetedForDeletion[otherFaceIdx] = true

	// f0VertKey := is.toVertKey(is.faces[f0])
	// f1VertKey := is.toVertKey(is.faces[f1])
	// if f0VertKey == f1VertKey {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces are identical! f0VertKey=%v", edge, f0VertKey)
	// } else {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces DIFFER!\nf0VertKey=%v\nf1VertKey=%v", edge, f0VertKey, f1VertKey)
	// }
}

func (fi *faceInfoT) connectOpenSrcExtrusionsToDst() {
	edgeLoops := fi.src.badEdgesToConnectedEdgeLoops()
	for faceStr, vertIndices := range edgeLoops {
		if deleteFaceIdx, ok := fi.dst.faceStrToFaceIdx[faceStr]; ok {
			fi.dst.facesTargetedForDeletion[deleteFaceIdx] = true
			continue
		}

		// Using this imaginary face "signature", find a dst face that shares one edge
		// and has the inverse normal to this face, then cut it and modify its neighbors.
		var matchingEdges []edgeT
		for i, vertIdx := range vertIndices {
			nextIdx := vertIndices[(i+1)%len(vertIndices)]
			edge := makeEdge(vertIdx, nextIdx)

		}

		log.Printf("WARNING: connectOpenSrcExtrusionsToDst: dst face not found: %v", faceStr)
	}
}

type edgeLoopT struct {
	face FaceT
}

func (el *edgeLoopT) addVertIdx(vIdx VertIndexT) {
	for _, vertIdx := range el.face {
		if vertIdx == vIdx {
			return
		}
	}
	el.face = append(el.face, vIdx)
}

func (is *infoSetT) badEdgesToConnectedEdgeLoops() map[faceKeyT]FaceT {
	vertsToEdgeLoops := map[VertIndexT]*edgeLoopT{}
	edgeLoops := map[*edgeLoopT]*edgeLoopT{}
	newEdgeLoop := func(edge edgeT) {
		el := &edgeLoopT{face: FaceT{edge[0], edge[1]}}
		vertsToEdgeLoops[edge[0]] = el
		vertsToEdgeLoops[edge[1]] = el
		edgeLoops[el] = el
	}
	addEdgeToLoop := func(edge edgeT, el *edgeLoopT) {
		el.addVertIdx(edge[0])
		el.addVertIdx(edge[1])
		vertsToEdgeLoops[edge[0]] = el
		vertsToEdgeLoops[edge[1]] = el
	}
	mergeTwoEdgeLoopsWithEdge := func(edge edgeT, edgeLoop1, edgeLoop2 *edgeLoopT) {
		addEdgeToLoop(edge, edgeLoop1)
		for _, vertIdx := range edgeLoop2.face {
			edgeLoop1.addVertIdx(vertIdx)
			vertsToEdgeLoops[vertIdx] = edgeLoop1
		}
		delete(edgeLoops, edgeLoop2)
	}

	for edge := range is.badEdges {
		edgeLoop1, ok1 := vertsToEdgeLoops[edge[0]]
		edgeLoop2, ok2 := vertsToEdgeLoops[edge[1]]
		switch {
		case ok1 && ok2 && edgeLoop1 == edgeLoop2:
			addEdgeToLoop(edge, edgeLoop1)
		case ok1 && ok2: // && edgeLoop1!=edgeLoop2: - delete the one edge loop and merge into the other
			mergeTwoEdgeLoopsWithEdge(edge, edgeLoop1, edgeLoop2)
		case ok1:
			addEdgeToLoop(edge, edgeLoop1)
		case ok2:
			addEdgeToLoop(edge, edgeLoop2)
		default:
			newEdgeLoop(edge)
		}
	}

	result := make(map[faceKeyT]FaceT, len(edgeLoops))
	for _, edgeLoop := range edgeLoops {
		result[edgeLoop.face.toKey()] = edgeLoop.face
	}

	return result
}
