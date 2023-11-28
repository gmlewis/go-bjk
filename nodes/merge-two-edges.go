// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"
	"sort"

	"golang.org/x/exp/maps"
)

func (fi *faceInfoT) merge2manis2edgesSrcFewerThanDst(sharedEdges sharedEdgesMapT, srcFaceIndicesToEdges, dstFaceIndicesToEdges face2EdgesMapT) {
	// if len(sharedEdges) != 2 {
	// 	log.Fatalf("merge2manis2edgesSrcFewerThanDst: sharedEdges=%v, want 2", len(sharedEdges))
	// }

	// debug
	edgeKeys := maps.Keys(sharedEdges)
	slices.SortFunc(edgeKeys, cmpEdges) // helps with debugging
	{
		log.Printf("\n\nmerge2manis2edgesSrcFewerThanDst: edges=%+v", edgeKeys)
		for i, edge := range edgeKeys {
			log.Printf("merge2manis2edgesSrcFewerThanDst: edge #%v of %v: %v, src faces: %+v, dst faces: %+v", i+1, len(edgeKeys), edge, sharedEdges[edge][0], sharedEdges[edge][1])
		}
		log.Printf("merge2manis2edgesSrcFewerThanDst: #srcFaceIndicesToEdges=%v, #dstFaceIndicesToEdges=%v", len(srcFaceIndicesToEdges), len(dstFaceIndicesToEdges))
		srcFaceIndices := maps.Keys(srcFaceIndicesToEdges)
		sort.Slice(srcFaceIndices, func(a, b int) bool { return srcFaceIndices[a] < srcFaceIndices[b] })
		for i, srcFaceIdx := range srcFaceIndices {
			log.Printf("merge2manis2edgesSrcFewerThanDst: src face #%v of %v: %v", i+1, len(srcFaceIndices), fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
		}
		dstFaceIndices := maps.Keys(dstFaceIndicesToEdges)
		sort.Slice(dstFaceIndices, func(a, b int) bool { return dstFaceIndices[a] < dstFaceIndices[b] })
		for i, dstFaceIdx := range dstFaceIndices {
			log.Printf("merge2manis2edgesSrcFewerThanDst: dst face #%v of %v: %v", i+1, len(dstFaceIndices), fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
		}
	}

	edgesToAbuttedFaces := fi.abuttedFaces(sharedEdges)

	for i, edge := range edgeKeys {
		log.Printf("merge2manis2edgesSrcFewerThanDst: edge #%v of %v: %v, %v abutted faces:", i+1, len(edgeKeys), edge, len(edgesToAbuttedFaces[edge]))
		for srcNormalKey, faceIndices := range edgesToAbuttedFaces[edge] {
			srcFaceIndices := faceIndices[0]
			dstFaceIndices := faceIndices[1]
			for j, srcFaceIdx := range srcFaceIndices {
				log.Printf("merge2manis2edgesSrcFewerThanDst: edge %v (normal key %v) abutted src face #%v of %v: %v", edge, srcNormalKey, j+1, len(srcFaceIndices), fi.m.dumpFace(srcFaceIdx, fi.src.faces[srcFaceIdx]))
			}
			for j, dstFaceIdx := range dstFaceIndices {
				log.Printf("merge2manis2edgesDstFewerThanDst: edge %v (normal key %v) abutted dst face #%v of %v: %v", edge, srcNormalKey, j+1, len(dstFaceIndices), fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
			}

			if len(srcFaceIndices) != 1 || len(dstFaceIndices) != 1 {
				log.Printf("WARNING: merge2manis2edgesDstFewerThanDst: unhandled case: edge #%v of %v: %v (normal key %v) has %v src faces and %v dst faces (want 1 and 1)", i+1, len(edgeKeys), edge, srcNormalKey, len(srcFaceIndices), len(dstFaceIndices))
				continue
			}

			srcFaceIdx, dstFaceIdx := srcFaceIndices[0], dstFaceIndices[0]
			fi.mergeAbuttedFacesOnEdge(edge, srcFaceIdx, dstFaceIdx)
		}
	}
}

func (fi *faceInfoT) mergeAbuttedFacesOnEdge(edge edgeT, srcFaceIdx, dstFaceIdx faceIndexT) {
	if fi.src.facesTargetedForDeletion[srcFaceIdx] {
		log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v already deleted - skipping", srcFaceIdx)
		return
	}
	if fi.dst.facesTargetedForDeletion[dstFaceIdx] {
		log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v already deleted - skipping", dstFaceIdx)
		return
	}

	srcEVs := fi.src.makeEdgeVectors(edge, srcFaceIdx)
	dstEVs := fi.dst.makeEdgeVectors(edge, dstFaceIdx)
	log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v: EVs[0] = %v", srcFaceIdx, srcEVs[0])
	log.Printf("mergeAbuttedFacesOnEdge: srcFaceIdx=%v: EVs[1] = %v", srcFaceIdx, srcEVs[1])
	log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v: EVs[0] = %v", dstFaceIdx, dstEVs[0])
	log.Printf("mergeAbuttedFacesOnEdge: dstFaceIdx=%v: EVs[1] = %v", dstFaceIdx, dstEVs[1])

	srcLength := (srcEVs[0].length + srcEVs[0].length) / 2
	dstLength := (dstEVs[0].length + dstEVs[0].length) / 2
	if srcLength < dstLength {
		fi.src.facesTargetedForDeletion[srcFaceIdx] = true
		fi.dst.resizeFace(dstFaceIdx, dstEVs[0].edge, dstEVs[1].edge, srcEVs) // resize dst by shorter edge vectors
	} else {
		fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
		fi.src.resizeFace(srcFaceIdx, srcEVs[0].edge, srcEVs[1].edge, dstEVs) // resize src by shorter edge vectors
	}
}

func (is *infoSetT) resizeFace(faceIdx faceIndexT, affectedEdge0, affectedEdge1 edgeT, evs [2]edgeVectorT) {
	face := is.faces[faceIdx]
	for i, vIdx := range face {
		switch {
		case vIdx == evs[0].fromVertIdx:
			log.Printf("resizeFace(faceIdx=%v) moving vertIdx %v to %v", faceIdx, vIdx, evs[0].toVertIdx)
			face[i] = evs[0].toVertIdx
		case vIdx == evs[1].fromVertIdx:
			log.Printf("resizeFace(faceIdx=%v) moving vertIdx %v to %v", faceIdx, vIdx, evs[1].toVertIdx)
			face[i] = evs[1].toVertIdx
		}
	}

	// now handle all the affected edges
	handle := func(edge edgeT) {
		for _, fIdx := range is.edgeToFaces[edge] {
			if fIdx == faceIdx || is.facesTargetedForDeletion[fIdx] {
				continue
			}
			is.insertVertOnEdge(fIdx, edge, evs)
		}
	}
	handle(affectedEdge0)
	handle(affectedEdge1)
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
			log.Printf("insertVertOnEdge(faceIdx=%v, edge=%v): inserting vertIdx=%v at position %v", faceIdx, edge, idx, nextI)
			log.Printf("insertVertOnEdge: result: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
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

type abutMapT map[edgeT]map[vertKeyT][2][]faceIndexT

// abuttedFaces returns a map from edge to normal key to two slices (src,dst) of abutted face indices.
// It sortes the indices to make testing easier.
func (fi *faceInfoT) abuttedFaces(sharedEdges sharedEdgesMapT) abutMapT {
	result := abutMapT{}
	addResult := func(edge edgeT, srcNormalKey vertKeyT, srcFaces []faceIndexT, dstFaceIdx faceIndexT) {
		nm, ok := result[edge]
		if !ok {
			nm = map[vertKeyT][2][]faceIndexT{}
			result[edge] = nm
		}
		vs, ok := nm[srcNormalKey]
		if !ok {
			vs = [2][]faceIndexT{srcFaces, {dstFaceIdx}}
			nm[srcNormalKey] = vs
			return
		}
		vs[1] = append(vs[1], dstFaceIdx)
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
				addResult(edge, key, srcFaceIndices, dstFaceIdx)
			}
		}
	}
	return result
}
