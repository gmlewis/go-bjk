package nodes

import (
	"log"

	"golang.org/x/exp/maps"
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

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges-1]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges+1 {
		// two separate cuts that need to be performed...
		srcMainFaceIdx, dstMainFaceIdx := srcEdgeCountToFaceIndices[numSharedEdges][0], dstEdgeCountToFaceIndices[numSharedEdges-1][0]
		firstCutSharedEdges := make(sharedEdgesMapT, numSharedEdges-1)
		secondCutSharedEdges := make(sharedEdgesMapT, 1)
		var secondCutDstMainFaceIdx faceIndexT
		for _, edge := range srcFaceIndicesToEdges[srcMainFaceIdx] {
			if v, ok := sharedEdges[edge]; ok && (v[1][0] == dstMainFaceIdx || v[1][1] == dstMainFaceIdx) {
				firstCutSharedEdges[edge] = sharedEdges[edge]
			} else {
				secondCutSharedEdges[edge] = sharedEdges[edge]
				log.Printf("fi.dst.faceNormals[%v]=%v", dstMainFaceIdx, fi.dst.faceNormals[dstMainFaceIdx])
				log.Printf("fi.dst.faceNormals[%v]=%v", v[1][0], fi.dst.faceNormals[v[1][0]])
				log.Printf("fi.dst.faceNormals[%v]=%v", v[1][1], fi.dst.faceNormals[v[1][1]])
				if fi.dst.faceNormals[v[1][0]].AboutEq(fi.dst.faceNormals[dstMainFaceIdx]) {
					secondCutDstMainFaceIdx = v[1][0]
					log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
				}
				if fi.dst.faceNormals[v[1][1]].AboutEq(fi.dst.faceNormals[dstMainFaceIdx]) {
					secondCutDstMainFaceIdx = v[1][1]
					log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
				}
			}
		}
		fi.merge2manisManyEdgesTwoFaces(firstCutSharedEdges, srcMainFaceIdx, dstMainFaceIdx)

		fi.merge2manisManyEdgesTwoFaces(secondCutSharedEdges, srcMainFaceIdx, secondCutDstMainFaceIdx)
		return
	}

	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: numSharedEdges=%v", numSharedEdges)
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: sharedEdges keys=%v=%+v", len(sharedEdges), maps.Keys(sharedEdges))
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices=%v=%+v", len(srcEdgeCountToFaceIndices), srcEdgeCountToFaceIndices)
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[%v]=%v=%+v", numSharedEdges, len(srcEdgeCountToFaceIndices[numSharedEdges]), srcEdgeCountToFaceIndices[numSharedEdges])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[%v]=%v=%+v", numSharedEdges-1, len(srcEdgeCountToFaceIndices[numSharedEdges-1]), srcEdgeCountToFaceIndices[numSharedEdges-1])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[2]=%v=%+v", len(srcEdgeCountToFaceIndices[2]), srcEdgeCountToFaceIndices[2])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[1]=%v=%+v", len(srcEdgeCountToFaceIndices[1]), srcEdgeCountToFaceIndices[1])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices=%v=%+v", len(dstEdgeCountToFaceIndices), dstEdgeCountToFaceIndices)
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[%v]=%v=%+v", numSharedEdges, len(dstEdgeCountToFaceIndices[numSharedEdges]), dstEdgeCountToFaceIndices[numSharedEdges])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[%v]=%v=%+v", numSharedEdges-1, len(dstEdgeCountToFaceIndices[numSharedEdges-1]), dstEdgeCountToFaceIndices[numSharedEdges-1])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[2]=%v=%+v", len(dstEdgeCountToFaceIndices[2]), dstEdgeCountToFaceIndices[2])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[1]=%v=%+v", len(dstEdgeCountToFaceIndices[1]), dstEdgeCountToFaceIndices[1])
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcFaceIndicesToEdges=%v", len(srcFaceIndicesToEdges))
	for faceIdx, edges := range srcFaceIndicesToEdges {
		log.Printf("srcFaceIndicesToEdges[%v]=%v=%+v", faceIdx, len(edges), edges)
	}
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstFaceIndicesToEdges=%v", len(dstFaceIndicesToEdges))
	for faceIdx, edges := range dstFaceIndicesToEdges {
		log.Printf("dstFaceIndicesToEdges[%v]=%v=%+v", faceIdx, len(edges), edges)
	}
}

/*
2023/11/16 17:28:54 manifoldMerge: src.badEdges=0=map[]
2023/11/16 17:28:54 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 17:28:54 merge2manifolds: shared verts: map[154:[[0 1 9] [81 94 95]] 155:[[0 1 2] [81 82 95]] 156:[[0 2 3] [82 83 95]] 157:[[0 3 4] [83 84 95]] 158:[[0 4 5] [84 85 95]] 159:[[0 5 6] [85 86 95]] 160:[[0 6 7] [86 87 95]] 192:[[0 7 8] [111 112 114]] 193:[[0 8 9] [112 113 114]]]
2023/11/16 17:28:54 merge2manifolds: shared edges: map[[154 155]:[[0 1] [81 95]] [155 156]:[[0 2] [82 95]] [156 157]:[[0 3] [83 95]] [157 158]:[[0 4] [84 95]] [158 159]:[[0 5] [85 95]] [159 160]:[[0 6] [86 95]] [192 193]:[[0 8] [112 114]]]
2023/11/16 17:28:54 merge2manifolds: shared faces: map[]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: numSharedEdges=7
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: sharedEdges keys=7=[[157 158] [154 155] [156 157] [158 159] [159 160] [192 193] [155 156]]

2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices=2=map[1:[1 3 5 6 8 2 4] 7:[0]]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[7]=1=[0]  // MAIN FACE 0
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[6]=0=[]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[2]=0=[]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[1]=7=[1 3 5 6 8 2 4]

2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices=2=map[1:[85 81 84 83 86 112 114 82] 6:[95]]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[7]=0=[]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[6]=1=[95]  // MAIN FACE 95
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[2]=0=[]
2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[1]=8=[85 81 84 83 86 112 114 82]

2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: srcFaceIndicesToEdges=8
2023/11/16 17:28:54 srcFaceIndicesToEdges[1]=1=[[154 155]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[0]=7=[[156 157] [158 159] [159 160] [192 193] [155 156] [157 158] [154 155]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[3]=1=[[156 157]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[5]=1=[[158 159]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[6]=1=[[159 160]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[8]=1=[[192 193]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[2]=1=[[155 156]]
2023/11/16 17:28:54 srcFaceIndicesToEdges[4]=1=[[157 158]]

2023/11/16 17:28:54 WARNING: merge2manisManyEdges: not implemented yet: dstFaceIndicesToEdges=9
2023/11/16 17:28:54 dstFaceIndicesToEdges[82]=1=[[155 156]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[84]=1=[[157 158]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[83]=1=[[156 157]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[95]=6=[[156 157] [158 159] [159 160] [155 156] [157 158] [154 155]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[86]=1=[[159 160]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[112]=1=[[192 193]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[114]=1=[[192 193]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[85]=1=[[158 159]]
2023/11/16 17:28:54 dstFaceIndicesToEdges[81]=1=[[154 155]]
*/

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
	newCutFaceOKToAdd := func(f FaceT) bool {
		edge := makeEdge(f[0], f[1])
		_, ok := sharedEdges[edge]
		return !ok
	}

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

		srcOtherEV := fi.src.connectedEdgeVectorFromVertOnFace(vertIdx, edge, srcOtherFaceIdx)
		srcEdgeVector := srcOtherEV.toSubFrom
		// log.Printf("merge2manisManyEdgesTwoFaces: edge %v, srcOtherVertIdx=%v, srcEdgeVector=%v", edge, srcOtherVertIdx, srcEdgeVector)

		dstOtherEV := fi.dst.connectedEdgeVectorFromVertOnFace(vertIdx, edge, dstOtherFaceIdx)
		dstEdgeVector := dstOtherEV.toSubFrom
		// log.Printf("merge2manisManyEdgesTwoFaces: edge %v, dstOtherVertIdx=%v, dstEdgeVector=%v", edge, dstOtherVertIdx, dstEdgeVector)

		srcEdgeLength := srcEdgeVector.Length()
		dstEdgeLength := dstEdgeVector.Length()
		switch {
		case AboutEq(srcEdgeLength, dstEdgeLength):
			fi.src.facesTargetedForDeletion[srcOtherFaceIdx] = true
			fi.dst.facesTargetedForDeletion[dstOtherFaceIdx] = true
		case srcEdgeLength < dstEdgeLength:
			fi.src.facesTargetedForDeletion[srcOtherFaceIdx] = true
			// Note that the first cutNeighborsAndShortenFaceOnEdge will affect ALL the edges!
			// Therefore, first check to make sure the dstOtherFaceIdx still has the original edge first.
			if cutFunc == nil {
				cutFunc = func() {
					fi.dst.cutNeighborsAndShortenFaceOnEdge(dstMainFaceIdx, srcEdgeVector, edge, newCutFaceOKToAdd)
				}
			}
		default: // srcEdgeLength > dstEdgeLength:
			fi.dst.facesTargetedForDeletion[dstOtherFaceIdx] = true
			// Same here. Check that srcOtherFaceIdx still has the original edge first.
			if cutFunc == nil {
				cutFunc = func() {
					fi.src.cutNeighborsAndShortenFaceOnEdge(srcMainFaceIdx, dstEdgeVector, edge, newCutFaceOKToAdd)
				}
			}
		}
	}

	if cutFunc != nil {
		cutFunc()
	}
}
