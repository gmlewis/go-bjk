package nodes

import (
	"log"
)

func (fi *faceInfoT) merge2manisManyEdges(sharedEdges sharedEdgesMapT) {
	// check if all src edges belong to a single face and if all dst edges belong to a single face.
	numSharedEdges := len(sharedEdges)
	srcFaceIndicesToEdges, dstFaceIndicesToEdges := reverseMapFaceIndicesToEdges(sharedEdges)

	srcEdgeCountToFaceIndices := faceIndicesByEdgeCount(srcFaceIndicesToEdges)
	dstEdgeCountToFaceIndices := faceIndicesByEdgeCount(dstFaceIndicesToEdges)

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges {
		srcMainFaceIdx, dstMainFaceIdx := srcEdgeCountToFaceIndices[numSharedEdges][0], dstEdgeCountToFaceIndices[numSharedEdges][0]
		fi.merge2manisManyEdgesTwoFaces(sharedEdges, srcMainFaceIdx, dstMainFaceIdx)
		return
	}

	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcFaceIndicesToEdges=%+v, dstFaceIndicesToEdges=%+v, srcEdgeCountToFaceIndices=%+v, dstEdgeCountToFaceIndices=%+v", srcFaceIndicesToEdges, dstFaceIndicesToEdges, srcEdgeCountToFaceIndices, dstEdgeCountToFaceIndices)
}

func (fi *faceInfoT) merge2manisManyEdgesTwoFaces(sharedEdges sharedEdgesMapT, srcMainFaceIdx, dstMainFaceIdx faceIndexT) {
	log.Printf("merge2manisManyEdgesTwoFaces: srcMainFaceIdx=%v", srcMainFaceIdx)
	log.Printf("merge2manisManyEdgesTwoFaces: dstMainFaceIdx=%v", dstMainFaceIdx)
	getOtherSrcFaceIndex := func(srcFaces []faceIndexT) faceIndexT {
		if srcFaces[0] == srcMainFaceIdx {
			return srcFaces[1]
		}
		return srcFaces[0]
	}
	getOtherDstFaceIndex := func(dstFaces []faceIndexT) faceIndexT {
		if dstFaces[0] == dstMainFaceIdx {
			return dstFaces[1]
		}
		return dstFaces[0]
	}

	// Once the neighbors are cut, all the other checks fail.
	// Therefore, save the function that will cut all the faces and run it after
	// all the other processing is finished.
	var cutFunc func()

	srcFacesToDeleteMap := map[faceIndexT]bool{}
	dstFacesToDeleteMap := map[faceIndexT]bool{}
	for edge, v := range sharedEdges {
		srcFaces := v[0]
		dstFaces := v[1]

		srcOtherFaceIdx := getOtherSrcFaceIndex(srcFaces)
		dstOtherFaceIdx := getOtherDstFaceIndex(dstFaces)

		if !fi.src.faceNormals[srcOtherFaceIdx].AboutEq(fi.dst.faceNormals[dstOtherFaceIdx].Negated()) {
			log.Printf("WARNING! merge2manisManyEdgesTwoFaces: edge %v, src other face[%v] normal %v is not opposite dst other face[%v] normal %v", edge, srcOtherFaceIdx, fi.src.faceNormals[srcOtherFaceIdx], dstOtherFaceIdx, fi.dst.faceNormals[dstOtherFaceIdx])
			return
		}

		log.Printf("merge2manisManyEdgesTwoFaces: edge %v, src other %v", edge, fi.m.dumpFace(srcOtherFaceIdx, fi.src.faces[srcOtherFaceIdx]))
		log.Printf("merge2manisManyEdgesTwoFaces: edge %v, dst other %v", edge, fi.m.dumpFace(dstOtherFaceIdx, fi.dst.faces[dstOtherFaceIdx]))

		vertIdx := edge[0]
		srcOtherVertIdx, srcEdgeVector := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcOtherFaceIdx)
		log.Printf("merge2manisManyEdgesTwoFaces: edge %v, srcOtherVertIdx=%v, srcEdgeVector=%v", edge, srcOtherVertIdx, srcEdgeVector)
		dstOtherVertIdx, dstEdgeVector := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstOtherFaceIdx)
		log.Printf("merge2manisManyEdgesTwoFaces: edge %v, dstOtherVertIdx=%v, dstEdgeVector=%v", edge, dstOtherVertIdx, dstEdgeVector)

		srcEdgeLength := srcEdgeVector.Length()
		dstEdgeLength := dstEdgeVector.Length()
		switch {
		case AboutEq(srcEdgeLength, dstEdgeLength):
			srcFacesToDeleteMap[srcOtherFaceIdx] = true
			dstFacesToDeleteMap[dstOtherFaceIdx] = true
		case srcEdgeLength < dstEdgeLength:
			srcFacesToDeleteMap[srcOtherFaceIdx] = true
			// Note that the first cutNeighborsAndShortenFaceOnEdge will affect ALL the edges!
			// Therefore, first check to make sure the dstOtherFaceIdx still has the original edge first.
			if cutFunc == nil {
				cutFunc = func() {
					fi.dst.cutNeighborsAndShortenFaceOnEdge(dstMainFaceIdx, srcEdgeVector, edge)
				}
			}
		default: // srcEdgeLength > dstEdgeLength:
			dstFacesToDeleteMap[dstOtherFaceIdx] = true
			// Same here. Check that srcOtherFaceIdx still has the original edge first.
			if cutFunc == nil {
				cutFunc = func() {
					fi.src.cutNeighborsAndShortenFaceOnEdge(srcMainFaceIdx, dstEdgeVector, edge)
				}
			}
		}
	}

	if cutFunc != nil {
		cutFunc()
	}

	fi.src.deleteFacesLastToFirst(srcFacesToDeleteMap)
	fi.dst.deleteFacesLastToFirst(dstFacesToDeleteMap)
}
