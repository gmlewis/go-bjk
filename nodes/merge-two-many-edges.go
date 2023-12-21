// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"

	"golang.org/x/exp/maps"
)

func (fi *faceInfoT) merge2manisManyEdges(sharedEdges sharedEdgesMapT) {
	// check if all src edges belong to a single face and if all dst edges belong to a single face.
	numSharedEdges := len(sharedEdges)
	srcFaceIndicesToEdges, dstFaceIndicesToEdges := reverseMapFaceIndicesToEdges(sharedEdges)

	// // debug
	// {
	// 	log.Printf("\n\nmerge2manisManyEdges: numSharedEdges=%v, #srcFaceIndicesToEdges=%v, #dstFaceIndicesToEdges=%v", numSharedEdges, len(srcFaceIndicesToEdges), len(dstFaceIndicesToEdges))
	// 	srcFaceIndices := maps.Keys(srcFaceIndicesToEdges)
	// 	sort.Slice(srcFaceIndices, func(a, b int) bool { return srcFaceIndices[a] < srcFaceIndices[b] })
	// 	for i, srcFaceIdx := range srcFaceIndices {
	// 		log.Printf("merge2manisManyEdges: src face #%v of %v: %v", i+1, len(srcFaceIndices), fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
	// 	}
	// 	dstFaceIndices := maps.Keys(dstFaceIndicesToEdges)
	// 	sort.Slice(dstFaceIndices, func(a, b int) bool { return dstFaceIndices[a] < dstFaceIndices[b] })
	// 	for i, dstFaceIdx := range dstFaceIndices {
	// 		log.Printf("merge2manisManyEdges: dst face #%v of %v: %v", i+1, len(dstFaceIndices), fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
	// 	}
	// }

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
	// log.Printf("merge2manisManyEdges: #srcEdgeCountToFaceIndices=%v, #dstEdgeCountToFaceIndices=%v", len(srcEdgeCountToFaceIndices), len(dstEdgeCountToFaceIndices))

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges {
		srcMainFaceIdx, dstMainFaceIdx := srcEdgeCountToFaceIndices[numSharedEdges][0], dstEdgeCountToFaceIndices[numSharedEdges][0]
		fi.merge2manisManyEdgesTwoFaces(sharedEdges, srcMainFaceIdx, dstMainFaceIdx)
		return
	}

	if len(srcEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(dstEdgeCountToFaceIndices[numSharedEdges-1]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges+1 {
		// two separate cuts that need to be performed...
		opts.srcMainFaceIdx = opts.srcEdgeCountToFaceIndices[numSharedEdges][0]
		opts.dstMainFaceIdx = opts.dstEdgeCountToFaceIndices[numSharedEdges-1][0]
		fi.twoSeparateCuts(opts)
		return
	}

	// This is the mirror swapped dst<=>src case of the case above.
	if len(dstEdgeCountToFaceIndices[numSharedEdges]) == 1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges &&
		len(srcEdgeCountToFaceIndices[numSharedEdges-1]) == 1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges+1 {
		fi.swapSrcAndDst(&opts)
		// two separate cuts that need to be performed...
		opts.srcMainFaceIdx = opts.srcEdgeCountToFaceIndices[numSharedEdges][0]
		opts.dstMainFaceIdx = opts.dstEdgeCountToFaceIndices[numSharedEdges-1][0]
		fi.twoSeparateCuts(opts)
		return
	}

	if len(srcEdgeCountToFaceIndices[1]) == numSharedEdges-1 && len(dstEdgeCountToFaceIndices[1]) == numSharedEdges {
		if fi.findMatchingFaceNormals(&opts, false) {
			fi.twoSeparateCuts(opts)
			return
		}
	}

	// This is the mirror swapped dst<=>src case of the case above.
	if len(dstEdgeCountToFaceIndices[1]) == numSharedEdges-1 && len(srcEdgeCountToFaceIndices[1]) == numSharedEdges {
		fi.swapSrcAndDst(&opts)
		if fi.findMatchingFaceNormals(&opts, false) {
			fi.twoSeparateCuts(opts)
			return
		}
	}

	if len(srcFaceIndicesToEdges) > len(dstFaceIndicesToEdges) {
		fi.swapSrcAndDst(&opts)
	}

	if fi.findMatchingFaceNormals(&opts, true) {
		fi.src.facesTargetedForDeletion[opts.srcMainFaceIdx] = true
		fi.dst.facesTargetedForDeletion[opts.dstMainFaceIdx] = true
		opts.deleteSharedEdges(opts.srcFaceIndicesToEdges[opts.srcMainFaceIdx])
		// log.Printf("merge2manisManyEdges: findNegatedFaceNormals=true, opts.sharedEdges=%#v", opts.sharedEdges)
		for edge, faceIndices := range opts.sharedEdges {
			fi.cutAndShortenEdge(edge, faceIndices[0], faceIndices[1])
		}
		return
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

func cmpEdges(e1, e2 edgeT) int {
	if e1[0] == e2[0] {
		return int(e1[1] - e2[1])
	}
	return int(e1[0] - e2[0])
}

func (fi *faceInfoT) findMatchingFaceNormals(opts *twoSeparateCutsOpts, negated bool) bool {
	// log.Printf("findMatchingFaceNormals(negated=%v): finding main src and dst faces", negated)
	srcFaceNormals := map[faceIndexT]Vec3{}
	for _, faceIndices := range opts.srcEdgeCountToFaceIndices {
		if len(faceIndices) != 1 {
			continue
		}
		faceIdx := faceIndices[0]
		normal := fi.src.faceNormals[faceIdx]
		if negated {
			normal = normal.Negated()
		}
		srcFaceNormals[faceIdx] = normal
		// log.Printf("findMatchingFaceNormals: src face[%v] normal: %v", faceIdx, srcFaceNormals[faceIdx])
	}

	var foundMatchingNormals bool
dstLoop:
	for _, faceIndices := range opts.dstEdgeCountToFaceIndices {
		if len(faceIndices) != 1 {
			continue
		}
		faceIdx := faceIndices[0]
		dstFaceNormal := fi.dst.faceNormals[faceIdx]
		// log.Printf("findMatchingFaceNormals: dst face[%v] normal: %v", faceIdx, dstFaceNormal)
		for srcFaceIdx, n := range srcFaceNormals {
			if n.AboutEq(dstFaceNormal) {
				opts.srcMainFaceIdx = srcFaceIdx
				opts.dstMainFaceIdx = faceIdx
				foundMatchingNormals = true
				break dstLoop
			}
		}
	}

	return foundMatchingNormals
}

func (fi *faceInfoT) swapSrcAndDst(opts *twoSeparateCutsOpts) {
	fi.src, fi.dst = fi.dst, fi.src
	if opts != nil {
		opts.srcFaceIndicesToEdges, opts.dstFaceIndicesToEdges = opts.dstFaceIndicesToEdges, opts.srcFaceIndicesToEdges
		opts.srcEdgeCountToFaceIndices, opts.dstEdgeCountToFaceIndices = opts.dstEdgeCountToFaceIndices, opts.srcEdgeCountToFaceIndices
		// reverse the sharedEdges results
		for edge, v := range opts.sharedEdges {
			opts.sharedEdges[edge] = [2][]faceIndexT{v[1], v[0]}
		}
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

func (opts *twoSeparateCutsOpts) deleteSharedEdges(edges []edgeT) {
	// log.Printf("deleteSharedEdges(edges=%+v)", edges)
	// log.Printf("deleteSharedEdges - before:\n%#v", *opts)
	for _, edge := range edges {
		delete(opts.sharedEdges, edge)
	}
	// opts.numSharedEdges = len(opts.sharedEdges)
	// opts.srcFaceIndicesToEdges, opts.dstFaceIndicesToEdges = reverseMapFaceIndicesToEdges(opts.sharedEdges)
	// opts.srcEdgeCountToFaceIndices = faceIndicesByEdgeCount(opts.srcFaceIndicesToEdges)
	// opts.dstEdgeCountToFaceIndices = faceIndicesByEdgeCount(opts.dstFaceIndicesToEdges)
	// log.Printf("deleteSharedEdges - after:\n%#v", *opts)
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
			// log.Printf("fi.dst.faceNormals[%v]=%v", opts.dstMainFaceIdx, fi.dst.faceNormals[opts.dstMainFaceIdx])
			// log.Printf("fi.dst.faceNormals[%v]=%v", v[1][0], fi.dst.faceNormals[v[1][0]])
			// log.Printf("fi.dst.faceNormals[%v]=%v", v[1][1], fi.dst.faceNormals[v[1][1]])
			if fi.dst.faceNormals[v[1][0]].AboutEq(fi.dst.faceNormals[opts.dstMainFaceIdx]) {
				secondCutDstMainFaceIdx = v[1][0]
				// log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
			}
			if fi.dst.faceNormals[v[1][1]].AboutEq(fi.dst.faceNormals[opts.dstMainFaceIdx]) {
				secondCutDstMainFaceIdx = v[1][1]
				// log.Printf("SETTING secondCutDstMainFaceIdx=%v", secondCutDstMainFaceIdx)
			}
		}
	}
	fi.merge2manisManyEdgesTwoFaces(firstCutSharedEdges, opts.srcMainFaceIdx, opts.dstMainFaceIdx)

	fi.merge2manisManyEdgesTwoFaces(secondCutSharedEdges, opts.srcMainFaceIdx, secondCutDstMainFaceIdx)
}

func (fi *faceInfoT) merge2manisManyEdgesTwoFaces(sharedEdges sharedEdgesMapT, srcMainFaceIdx, dstMainFaceIdx faceIndexT) {
	// log.Printf("merge2manisManyEdgesTwoFaces: srcMainFaceIdx=%v", srcMainFaceIdx)
	// log.Printf("merge2manisManyEdgesTwoFaces: dstMainFaceIdx=%v", dstMainFaceIdx)
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
			log.Printf("WARNING: merge2manisManyEdgesTwoFaces: unhandled case: edge %v, src other face[%v] normal %v is not opposite dst other face[%v] normal %v", edge, srcOtherFaceIdx, fi.src.faceNormals[srcOtherFaceIdx], dstOtherFaceIdx, fi.dst.faceNormals[dstOtherFaceIdx])
			return
		}

		// log.Printf("merge2manisManyEdgesTwoFaces: edge %v, src other %v", edge, fi.m.dumpFace(srcOtherFaceIdx, fi.src.faces[srcOtherFaceIdx]))
		// log.Printf("merge2manisManyEdgesTwoFaces: edge %v, dst other %v", edge, fi.m.dumpFace(dstOtherFaceIdx, fi.dst.faces[dstOtherFaceIdx]))

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
