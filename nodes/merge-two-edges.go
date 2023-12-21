// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"

	"golang.org/x/exp/maps"
)

func (fi *faceInfoT) merge2manis2edgesSrcFewerThanDst(sharedEdges sharedEdgesMapT, srcFaceIndicesToEdges, dstFaceIndicesToEdges face2EdgesMapT) {

	edgeKeys := maps.Keys(sharedEdges)
	slices.SortFunc(edgeKeys, cmpEdges) // helps with debugging

	// // debug
	// {
	// 	log.Printf("\n\nmerge2manis2edgesSrcFewerThanDst: edges=%+v", edgeKeys)
	// 	for i, edge := range edgeKeys {
	// 		log.Printf("edge #%v of %v: %v, src faces: %+v, dst faces: %+v", i+1, len(edgeKeys), edge, sharedEdges[edge][0], sharedEdges[edge][1])
	// 	}
	// 	log.Printf("#srcFaceIndicesToEdges=%v, #dstFaceIndicesToEdges=%v\n\n", len(srcFaceIndicesToEdges), len(dstFaceIndicesToEdges))
	// 	srcFaceIndices := maps.Keys(srcFaceIndicesToEdges)
	// 	sort.Slice(srcFaceIndices, func(a, b int) bool { return srcFaceIndices[a] < srcFaceIndices[b] })
	// 	for i, srcFaceIdx := range srcFaceIndices {
	// 		log.Printf("src face #%v of %v: %v", i+1, len(srcFaceIndices), fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
	// 	}
	// 	log.Printf("\n")
	// 	dstFaceIndices := maps.Keys(dstFaceIndicesToEdges)
	// 	sort.Slice(dstFaceIndices, func(a, b int) bool { return dstFaceIndices[a] < dstFaceIndices[b] })
	// 	for i, dstFaceIdx := range dstFaceIndices {
	// 		log.Printf("dst face #%v of %v: %v", i+1, len(dstFaceIndices), fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
	// 	}
	// }

	abutInfo := fi.abuttedFaces(sharedEdges)

	for i, edge := range edgeKeys {
		// log.Printf("\n\nmerge2manis2edgesSrcFewerThanDst: edge #%v of %v: %v, %v abutted faces:", i+1, len(edgeKeys), edge, len(abutInfo.edgesToAbuttedFaces[edge]))
		if len(abutInfo.edgesToAbuttedFaces[edge]) == 0 {
			fi.checkCompleteOverlapOnEdge(edge, sharedEdges, srcFaceIndicesToEdges, dstFaceIndicesToEdges)
			continue
		}

		for srcNormalKey, faceIndices := range abutInfo.edgesToAbuttedFaces[edge] {
			srcFaceIndices := faceIndices[0]
			dstFaceIndices := faceIndices[1]
			// for j, srcFaceIdx := range srcFaceIndices {
			// 	log.Printf("edge %v (normal key %v) abutted src face #%v of %v: %v", edge, srcNormalKey, j+1, len(srcFaceIndices), fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
			// }
			// for j, dstFaceIdx := range dstFaceIndices {
			// 	log.Printf("edge %v (normal key %v) abutted dst face #%v of %v: %v", edge, srcNormalKey, j+1, len(dstFaceIndices), fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
			// }

			if len(srcFaceIndices) != 1 || len(dstFaceIndices) != 1 {
				log.Printf("WARNING: merge2manis2edgesDstFewerThanDst: unhandled case: edge #%v of %v: %v (normal key %v) has %v src faces and %v dst faces (want 1 and 1)", i+1, len(edgeKeys), edge, srcNormalKey, len(srcFaceIndices), len(dstFaceIndices))
				continue
			}

			srcFaceIdx, dstFaceIdx := srcFaceIndices[0], dstFaceIndices[0]
			abutInfo.mergeAbuttedFacesOnEdge(edge, srcFaceIdx, dstFaceIdx)
		}
	}
}

func (ai *abutInfoT) mergeAbuttedFacesOnEdge(edge edgeT, srcFaceIdx, dstFaceIdx faceIndexT) {
	fi := ai.fi
	if fi.src.facesTargetedForDeletion[srcFaceIdx] {
		// log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v already deleted - skipping", srcFaceIdx)
		return
	}
	if fi.dst.facesTargetedForDeletion[dstFaceIdx] {
		// log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v already deleted - skipping", dstFaceIdx)
		return
	}

	srcEVs := fi.src.makeEdgeVectors(edge, srcFaceIdx)
	dstEVs := fi.dst.makeEdgeVectors(edge, dstFaceIdx)
	// log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v: EVs[0] = %v", srcFaceIdx, srcEVs[0])
	// log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v: EVs[1] = %v", srcFaceIdx, srcEVs[1])
	// log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v: EVs[0] = %v", dstFaceIdx, dstEVs[0])
	// log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v: EVs[1] = %v", dstFaceIdx, dstEVs[1])

	// first pass: delete and resize faces
	srcLength := (srcEVs[0].length + srcEVs[0].length) / 2
	dstLength := (dstEVs[0].length + dstEVs[0].length) / 2
	if srcLength < dstLength {
		// log.Printf("mergeAbuttedFacesOnEdge: marking src face for DELETION: %v", fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
		fi.src.facesTargetedForDeletion[srcFaceIdx] = true
		fi.dst.resizeFace(ai.dstFaces, dstFaceIdx, []edgeT{dstEVs[0].edge, dstEVs[1].edge}, srcEVs) // resize dst by shorter edge vectors
	} else {
		// log.Printf("mergeAbuttedFacesOnEdge: marking dst face for DELETION: %v", fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
		fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
		fi.src.resizeFace(ai.srcFaces, srcFaceIdx, []edgeT{srcEVs[0].edge, srcEVs[1].edge}, dstEVs) // resize src by shorter edge vectors
	}
}

func (is *infoSetT) resizeFace(avoidFaces map[faceIndexT]bool, faceIdx faceIndexT, affectedEdges []edgeT, evs [2]edgeVectorT) {
	face := is.faces[faceIdx]
	// first pass - move verts
	for i, vIdx := range face {
		switch {
		case vIdx == evs[0].fromVertIdx:
			// log.Printf("resizeFace(faceIdx=%v) moving vertIdx %v to %v", faceIdx, vIdx, evs[0].toVertIdx)
			face[i] = evs[0].toVertIdx
		case vIdx == evs[1].fromVertIdx:
			// log.Printf("resizeFace(faceIdx=%v) moving vertIdx %v to %v", faceIdx, vIdx, evs[1].toVertIdx)
			face[i] = evs[1].toVertIdx
		}
	}
	// second pass - simplify face if any verts are duped
	newFace := make(FaceT, 0, len(face))
	for i, vIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if vIdx == nextIdx {
			continue
		}
		newFace = append(newFace, vIdx)
	}
	is.faces[faceIdx] = newFace

	// now handle affected edges
	for _, edge := range affectedEdges {
		for _, fIdx := range is.edgeToFaces[edge] {
			if fIdx == faceIdx || is.facesTargetedForDeletion[fIdx] || (avoidFaces != nil && avoidFaces[fIdx]) {
				continue
			}
			is.insertVertOnEdge(fIdx, edge, evs)
		}
	}
}

func (is *infoSetT) insertVertOnEdge(faceIdx faceIndexT, edge edgeT, evs [2]edgeVectorT) {
	face := is.faces[faceIdx]
	for i, vIdx := range face {
		nextI := (i + 1) % len(face)
		nextIdx := face[nextI]
		if makeEdge(vIdx, nextIdx) != edge {
			continue
		}
		f := func(idx VertIndexT) {
			is.faces[faceIdx] = slices.Insert(face, nextI, idx)
			// log.Printf("insertVertOnEdge(faceIdx=%v, edge=%v): inserting vertIdx=%v at position %v", faceIdx, edge, idx, nextI)
			// log.Printf("insertVertOnEdge: result: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
		}
		switch {
		case evs[0].fromVertIdx == vIdx, evs[0].fromVertIdx == nextIdx:
			f(evs[0].toVertIdx)
		case evs[0].toVertIdx == vIdx, evs[0].toVertIdx == nextIdx:
			f(evs[0].fromVertIdx)
		case evs[1].fromVertIdx == vIdx, evs[1].fromVertIdx == nextIdx:
			f(evs[1].toVertIdx)
		case evs[1].toVertIdx == vIdx, evs[1].toVertIdx == nextIdx:
			f(evs[1].fromVertIdx)
		default:
			log.Fatalf("insertVertOnEdge: programming error")
		}
		return
	}
}

type abutInfoT struct {
	fi                  *faceInfoT
	edgesToAbuttedFaces abutMapT
	srcFaces            map[faceIndexT]bool
	dstFaces            map[faceIndexT]bool
}

type abutMapT map[edgeT]map[vertKeyT][2][]faceIndexT

// abuttedFaces returns a map from edge to normal key to two slices (src,dst) of abutted face indices.
// It sorts the indices to make testing easier.
func (fi *faceInfoT) abuttedFaces(sharedEdges sharedEdgesMapT) *abutInfoT {
	ai := &abutInfoT{
		fi:                  fi,
		edgesToAbuttedFaces: abutMapT{},
		srcFaces:            map[faceIndexT]bool{},
		dstFaces:            map[faceIndexT]bool{},
	}

	for edge, v := range sharedEdges {
		srcNormals := map[vertKeyT][]faceIndexT{}
		for _, srcFaceIdx := range v[0] {
			key := fi.src.faceNormals[srcFaceIdx].toKey()
			srcNormals[key] = append(srcNormals[key], srcFaceIdx)
		}
		for _, dstFaceIdx := range v[1] {
			key := fi.dst.faceNormals[dstFaceIdx].Negated().toKey() // flip the normal
			if srcFaceIndices, ok := srcNormals[key]; ok {
				ai.addResult(edge, key, srcFaceIndices, dstFaceIdx)
			}
		}
	}
	return ai
}

func (ai *abutInfoT) addResult(edge edgeT, srcNormalKey vertKeyT, srcFaces []faceIndexT, dstFaceIdx faceIndexT) {
	for _, srcFaceIdx := range srcFaces {
		ai.srcFaces[srcFaceIdx] = true
	}
	ai.dstFaces[dstFaceIdx] = true

	nm, ok := ai.edgesToAbuttedFaces[edge]
	if !ok {
		nm = map[vertKeyT][2][]faceIndexT{}
		ai.edgesToAbuttedFaces[edge] = nm
	}
	vs, ok := nm[srcNormalKey]
	if !ok {
		vs = [2][]faceIndexT{srcFaces, {dstFaceIdx}}
		nm[srcNormalKey] = vs
		return
	}
	vs[1] = append(vs[1], dstFaceIdx)
}

func (fi *faceInfoT) checkCompleteOverlapOnEdge(edge edgeT, sharedEdges sharedEdgesMapT, srcFaceIndicesToEdges, dstFaceIndicesToEdges face2EdgesMapT) {
	srcFaceIdx1 := sharedEdges[edge][0][1]
	if srcFaceIdx0 := sharedEdges[edge][0][0]; fi.m.faceArea(fi.src.faces[srcFaceIdx0]) < fi.m.faceArea(fi.src.faces[srcFaceIdx1]) {
		srcFaceIdx1 = srcFaceIdx0
	}
	// log.Printf("\n\ncheckCompleteOverlapOnEdge: srcFaceIdx0=%v area=%v, srcFaceIdx1=%v area=%v", srcFaceIdx0, fi.m.faceArea(fi.src.faces[srcFaceIdx0]), srcFaceIdx1, fi.m.faceArea(fi.src.faces[srcFaceIdx1]))

	dstFaceIdx0 := sharedEdges[edge][1][0]
	if dstFaceIdx1 := sharedEdges[edge][1][1]; fi.m.faceArea(fi.dst.faces[dstFaceIdx0]) > fi.m.faceArea(fi.dst.faces[dstFaceIdx1]) {
		dstFaceIdx0 = dstFaceIdx1
	}
	// log.Printf("checkCompleteOverlapOnEdge: dstFaceIdx0=%v area=%v, dstFaceIdx1=%v area=%v", dstFaceIdx0, fi.m.faceArea(fi.dst.faces[dstFaceIdx0]), dstFaceIdx1, fi.m.faceArea(fi.dst.faces[dstFaceIdx1]))

	// log.Printf("targeting smaller srcFaceIdx1=%v for deletion", srcFaceIdx1)
	fi.src.facesTargetedForDeletion[srcFaceIdx1] = true

	moveSrcEVs := fi.dst.makeEdgeVectors(edge, dstFaceIdx0)
	// log.Printf("moveSrcEVs[0]=%v", moveSrcEVs[0])
	// log.Printf("moveSrcEVs[1]=%v", moveSrcEVs[1])
	moveDstEVs := fi.src.makeEdgeVectors(edge, srcFaceIdx1)
	// log.Printf("BEFORE: moveDstEVs[0]=%v", moveDstEVs[0])
	// log.Printf("BEFORE: moveDstEVs[1]=%v", moveDstEVs[1])

	moveMap := fi.src.moveAllVertsOnDeletedFace(srcFaceIdx1, moveSrcEVs)
	// log.Printf("moveMap: %+v", moveMap)
	moveDstEVs[0].fromVertIdx = moveMap[moveDstEVs[0].fromVertIdx]
	moveDstEVs[0].toVertIdx = moveMap[moveDstEVs[0].toVertIdx]
	moveDstEVs[1].fromVertIdx = moveMap[moveDstEVs[1].fromVertIdx]
	moveDstEVs[1].toVertIdx = moveMap[moveDstEVs[1].toVertIdx]
	// log.Printf("AFTER: moveDstEVs[0]=%v", moveDstEVs[0])
	// log.Printf("AFTER: moveDstEVs[1]=%v", moveDstEVs[1])

	dstOppositeEdge := makeEdge(moveSrcEVs[0].toVertIdx, moveSrcEVs[1].toVertIdx)
	// log.Printf("dstOppositeEdge=%v: %v %v", dstOppositeEdge, fi.m.Verts[dstOppositeEdge[0]], fi.m.Verts[dstOppositeEdge[1]])
	dstFaceToResizeIdx, affectedDstEdge0, affectedDstEdge1 := fi.dst.otherFaceOnEdge(dstOppositeEdge, dstFaceIdx0)

	fi.dst.resizeFace(nil, dstFaceToResizeIdx, []edgeT{affectedDstEdge0, affectedDstEdge1}, moveDstEVs)
}

func (is *infoSetT) moveAllVertsOnDeletedFace(baseFaceIdx faceIndexT, evs [2]edgeVectorT) vToVMap {
	baseFace := is.faces[baseFaceIdx]
	moveMap := is.moveVerts(baseFace, evs[0].toSubFrom)

	processedFaces := map[faceIndexT]bool{baseFaceIdx: true}
	for oldIdx := range moveMap {
		for _, faceIdx := range is.vertToFaces[oldIdx] {
			if processedFaces[faceIdx] {
				continue
			}
			processedFaces[faceIdx] = true
			face := is.faces[faceIdx]
			for i, vIdx := range face {
				if newIdx, ok := moveMap[vIdx]; ok {
					is.faces[faceIdx][i] = newIdx
				}
			}
		}
	}

	return moveMap
}
