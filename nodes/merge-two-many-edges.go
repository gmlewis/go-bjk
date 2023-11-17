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
	opts := twoSeparateCutsOpts{
		numSharedEdges:            numSharedEdges,
		sharedEdges:               sharedEdges,
		srcFaceIndicesToEdges:     srcFaceIndicesToEdges,
		dstFaceIndicesToEdges:     dstFaceIndicesToEdges,
		srcEdgeCountToFaceIndices: srcEdgeCountToFaceIndices,
		dstEdgeCountToFaceIndices: dstEdgeCountToFaceIndices,
	}

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges {
		srcMainFaceIdx, dstMainFaceIdx := srcEdgeCountToFaceIndices[numSharedEdges][0], dstEdgeCountToFaceIndices[numSharedEdges][0]
		fi.merge2manisManyEdgesTwoFaces(sharedEdges, srcMainFaceIdx, dstMainFaceIdx)
		return
	}

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges-1]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges+1 {
		// two separate cuts that need to be performed...
		opts.srcMainFaceIdx = srcEdgeCountToFaceIndices[numSharedEdges][0]
		opts.dstMainFaceIdx = dstEdgeCountToFaceIndices[numSharedEdges-1][0]
		fi.twoSeparateCuts(opts)
		return
	}

	if len(srcEdgeCountToFaceIndices[1]) == numSharedEdges-1 &&
		len(dstEdgeCountToFaceIndices[1]) == numSharedEdges {
		log.Printf("merge2manisManyEdges: finding main src and dst faces")
		srcFaceNormals := map[faceIndexT]Vec3{}
		for _, faceIndices := range srcEdgeCountToFaceIndices {
			if len(faceIndices) != 1 {
				continue
			}
			faceIdx := faceIndices[0]
			srcFaceNormals[faceIdx] = fi.src.faceNormals[faceIdx]
			log.Printf("merge2manisManyEdges: src face[%v] normal: %v", faceIdx, srcFaceNormals[faceIdx])
		}

		var foundMatchingNormals bool
	dstLoop:
		for _, faceIndices := range dstEdgeCountToFaceIndices {
			if len(faceIndices) != 1 {
				continue
			}
			faceIdx := faceIndices[0]
			dstFaceNormal := fi.dst.faceNormals[faceIdx]
			log.Printf("merge2manisManyEdges: dst face[%v] normal: %v", faceIdx, dstFaceNormal)
			for srcFaceIdx, n := range srcFaceNormals {
				if n.AboutEq(dstFaceNormal) {
					opts.srcMainFaceIdx = srcFaceIdx
					opts.dstMainFaceIdx = faceIdx
					foundMatchingNormals = true
					break dstLoop
				}
			}
		}

		// // swap src/dst and call twoSeparateCuts
		// fi.src, fi.dst = fi.dst, fi.src
		// // reverse the sharedEdges results
		// for edge, v := range sharedEdges {
		// 	sharedEdges[edge] = [2][]faceIndexT{v[1], v[0]}
		// }
		if foundMatchingNormals {
			fi.twoSeparateCuts(opts)
			return
		}
	}

	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: numSharedEdges=%v", numSharedEdges)
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: sharedEdges keys=%v=%+v", len(sharedEdges), maps.Keys(sharedEdges))
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices=%v=%+v", len(srcEdgeCountToFaceIndices), srcEdgeCountToFaceIndices)
	for edgeCount, faceIndices := range srcEdgeCountToFaceIndices {
		log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[%v]=%v=%+v", edgeCount, len(faceIndices), faceIndices)
	}
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices=%v=%+v", len(dstEdgeCountToFaceIndices), dstEdgeCountToFaceIndices)
	for edgeCount, faceIndices := range dstEdgeCountToFaceIndices {
		log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[%v]=%v=%+v", edgeCount, len(faceIndices), faceIndices)
	}
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: srcFaceIndicesToEdges=%v", len(srcFaceIndicesToEdges))
	for faceIdx, edges := range srcFaceIndicesToEdges {
		log.Printf("srcFaceIndicesToEdges[%v]=%v=%+v", faceIdx, len(edges), edges)
	}
	log.Printf("WARNING: merge2manisManyEdges: not implemented yet: dstFaceIndicesToEdges=%v", len(dstFaceIndicesToEdges))
	for faceIdx, edges := range dstFaceIndicesToEdges {
		log.Printf("dstFaceIndicesToEdges[%v]=%v=%+v", faceIdx, len(edges), edges)
	}
}

type twoSeparateCutsOpts struct {
	numSharedEdges            int
	sharedEdges               sharedEdgesMapT
	srcFaceIndicesToEdges     face2EdgesMapT
	dstFaceIndicesToEdges     face2EdgesMapT
	srcEdgeCountToFaceIndices map[int][]faceIndexT
	dstEdgeCountToFaceIndices map[int][]faceIndexT
	srcMainFaceIdx            faceIndexT
	dstMainFaceIdx            faceIndexT
}

func (fi *faceInfoT) twoSeparateCuts(opts twoSeparateCutsOpts) {
	firstCutSharedEdges := make(sharedEdgesMapT, opts.numSharedEdges-1)
	secondCutSharedEdges := make(sharedEdgesMapT, 1)
	var secondCutDstMainFaceIdx faceIndexT
	for _, edge := range opts.srcFaceIndicesToEdges[opts.srcMainFaceIdx] {
		if v, ok := opts.sharedEdges[edge]; ok && (v[1][0] == opts.dstMainFaceIdx || v[1][1] == opts.dstMainFaceIdx) {
			firstCutSharedEdges[edge] = opts.sharedEdges[edge]
		} else {
			secondCutSharedEdges[edge] = opts.sharedEdges[edge]
			log.Printf("fi.dst.faceNormals[%v]=%v", opts.dstMainFaceIdx, fi.dst.faceNormals[opts.dstMainFaceIdx])
			log.Printf("fi.dst.faceNormals[%v]=%v", v[1][0], fi.dst.faceNormals[v[1][0]])
			log.Printf("fi.dst.faceNormals[%v]=%v", v[1][1], fi.dst.faceNormals[v[1][1]])
			if fi.dst.faceNormals[v[1][0]].AboutEq(fi.dst.faceNormals[opts.dstMainFaceIdx]) {
				secondCutDstMainFaceIdx = v[1][0]
				log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
			}
			if fi.dst.faceNormals[v[1][1]].AboutEq(fi.dst.faceNormals[opts.dstMainFaceIdx]) {
				secondCutDstMainFaceIdx = v[1][1]
				log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
			}
		}
	}
	fi.merge2manisManyEdgesTwoFaces(firstCutSharedEdges, opts.srcMainFaceIdx, opts.dstMainFaceIdx)

	fi.merge2manisManyEdgesTwoFaces(secondCutSharedEdges, opts.srcMainFaceIdx, secondCutDstMainFaceIdx)
}

/*
2023/11/16 21:01:04 manifoldMerge: src.badEdges=0=map[]
2023/11/16 21:01:04 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 21:01:04 merge2manifolds: shared verts: map[70:[[0 1 9] [33 46 47]] 71:[[0 1 2] [33 34 47]] 72:[[0 2 3] [34 35 47]] 73:[[0 3 4] [35 36 47]] 74:[[0 4 5] [36 37 47]] 75:[[0 5 6] [37 38 47]] 76:[[0 6 7] [38 39 47]] 236:[[7 8 10] [145 146 149 150]] 237:[[8 9 10] [146 147 150 151]] 240:[[0 7 8] [149 150 152]] 241:[[0 8 9] [150 151 152]]]
2023/11/16 21:01:04 merge2manifolds: shared edges: map[[70 71]:[[0 1] [33 47]] [71 72]:[[0 2] [34 47]] [72 73]:[[0 3] [35 47]] [73 74]:[[0 4] [36 47]] [74 75]:[[0 5] [37 47]] [75 76]:[[0 6] [38 47]] [236 237]:[[8 10] [146 150]] [236 240]:[[7 8] [149 150]] [237 241]:[[8 9] [150 151]] [240 241]:[[0 8] [150 152]]]
2023/11/16 21:01:04 merge2manifolds: shared faces: map[[236 237 240 241]:[8 150]]

2023/11/16 21:01:04 WARNING: merge2manisManyEdges: not implemented yet: numSharedEdges=10
2023/11/16 21:01:04 WARNING: merge2manisManyEdges: not implemented yet: sharedEdges keys=10=[[73 74] [237 241] [240 241] [71 72] [70 71] [236 240] [72 73] [75 76] [236 237] [74 75]]
2023/11/16 21:01:04 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices=3=map[1:[2 1 7 5 4 6 10 3 9] 4:[8] 7:[0]]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[7]=1=[0]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[1]=9=[4 1 9 6 2 3 5 7 10]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: srcEdgeCountToFaceIndices[4]=1=[8]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices=3=map[1:[33 35 38 146 152 37 151 149 34 36] 4:[150] 6:[47]]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[1]=10=[33 35 38 146 152 37 151 149 34 36]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[6]=1=[47]
2023/11/16 21:05:00 WARNING: merge2manisManyEdges: not implemented yet: dstEdgeCountToFaceIndices[4]=1=[150]
2023/11/16 21:01:04 WARNING: merge2manisManyEdges: not implemented yet: srcFaceIndicesToEdges=11

2023/11/16 21:01:04 srcFaceIndicesToEdges[8]=4=[[237 241] [240 241] [236 240] [236 237]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[2]=1=[[71 72]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[1]=1=[[70 71]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[7]=1=[[236 240]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[5]=1=[[74 75]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[9]=1=[[237 241]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[0]=7=[[240 241] [71 72] [70 71] [73 74] [75 76] [74 75] [72 73]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[4]=1=[[73 74]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[6]=1=[[75 76]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[10]=1=[[236 237]]
2023/11/16 21:01:04 srcFaceIndicesToEdges[3]=1=[[72 73]]
2023/11/16 21:01:04 WARNING: merge2manisManyEdges: not implemented yet: dstFaceIndicesToEdges=12
2023/11/16 21:01:04 dstFaceIndicesToEdges[37]=1=[[74 75]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[35]=1=[[72 73]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[33]=1=[[70 71]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[36]=1=[[73 74]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[146]=1=[[236 237]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[34]=1=[[71 72]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[47]=6=[[71 72] [70 71] [73 74] [75 76] [74 75] [72 73]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[149]=1=[[236 240]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[38]=1=[[75 76]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[150]=4=[[237 241] [240 241] [236 240] [236 237]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[151]=1=[[237 241]]
2023/11/16 21:01:04 dstFaceIndicesToEdges[152]=1=[[240 241]]
*/

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
