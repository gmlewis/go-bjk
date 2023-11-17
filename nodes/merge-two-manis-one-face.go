package nodes

import "log"

func (fi *faceInfoT) merge2manisOneFace(sharedEdges sharedEdgesMapT, srcFaceIdx, dstFaceIdx faceIndexT) {
	srcFaceNumVerts := len(fi.src.faces[srcFaceIdx])
	dstFaceNumVerts := len(fi.dst.faces[dstFaceIdx])
	if srcFaceNumVerts == dstFaceNumVerts && len(sharedEdges) == srcFaceNumVerts {
		fi.mergeExtrusion(sharedEdges, srcFaceIdx, dstFaceIdx)
		return
	}

	// this is not an extrusion afterall.
	fi.merge2manisManyEdges(sharedEdges)
}

func (fi *faceInfoT) mergeExtrusion(sharedEdges sharedEdgesMapT, srcFaceIdx, dstFaceIdx faceIndexT) {
	// log.Printf("mergeExtrusion: sharedEdges=%+v", sharedEdges)
	// log.Printf("mergeExtrusion: srcFaceIdx=%v, dstFaceIdx=%v", srcFaceIdx, dstFaceIdx)

	srcFaceNumVerts, dstFaceNumVerts := len(fi.src.faces[srcFaceIdx]), len(fi.dst.faces[dstFaceIdx])
	if srcFaceNumVerts == dstFaceNumVerts && srcFaceNumVerts == len(sharedEdges) &&
		fi.src.faceNormals[srcFaceIdx].AboutEq(fi.dst.faceNormals[dstFaceIdx].Negated()) {
		// abutting faces - remove them both.
		fi.src.facesTargetedForDeletion[srcFaceIdx] = true
		fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
		return
	}

	srcSideEVs := fi.src.getFaceSideEdgeVectors(srcFaceIdx)
	srcOtherEndFace := make(FaceT, 0, len(srcSideEVs))
	for i, ev := range srcSideEVs {
		nextEV := srcSideEVs[(i+1)%len(srcSideEVs)]
		if !ev.toSubFrom.AboutEq(nextEV.toSubFrom) {
			log.Printf("WARNING: mergeExtrusion: unhandled case: src ev=%+v, nextEV=%+v", ev, nextEV)
			return
		}
		srcOtherEndFace = append(srcOtherEndFace, ev.toVertIdx)
	}

	dstSideEVs := fi.dst.getFaceSideEdgeVectors(dstFaceIdx)
	dstOtherEndFace := make(FaceT, 0, len(dstSideEVs))
	for i, ev := range dstSideEVs {
		nextEV := dstSideEVs[(i+1)%len(dstSideEVs)]
		if !ev.toSubFrom.AboutEq(nextEV.toSubFrom) {
			log.Printf("WARNING: mergeExtrusion: unhandled case: dst ev=%+v, nextEV=%+v", ev, nextEV)
			return
		}
		dstOtherEndFace = append(dstOtherEndFace, ev.toVertIdx)
	}

	// all edgeVectors for src are identical and all EVs for dst are also identical. Find out which is longer and truncate it.
	if srcSideEVs[0].length > dstSideEVs[0].length {
		fi.truncateExtrusion(fi.src, srcSideEVs, dstSideEVs)
		fi.src.facesTargetedForDeletion[srcFaceIdx] = true
		dstFaceToDeleteIdx, ok := fi.dst.faceStr2FaceIdx[dstOtherEndFace.toKey()]
		if !ok {
			log.Fatalf("mergeExtrusion: unable to get dstFace to delete from %+v", dstOtherEndFace)
		}
		fi.dst.facesTargetedForDeletion[dstFaceToDeleteIdx] = true
		return
	}

	fi.truncateExtrusion(fi.dst, dstSideEVs, srcSideEVs)
	fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
	srcFaceToDeleteIdx, ok := fi.src.faceStr2FaceIdx[srcOtherEndFace.toKey()]
	if !ok {
		log.Fatalf("mergeExtrusion: unable to get srcFace to delete from %+v", srcOtherEndFace)
	}
	fi.src.facesTargetedForDeletion[srcFaceToDeleteIdx] = true
}

func (fi *faceInfoT) truncateExtrusion(is *infoSetT, evsToTruncate, otherEVs []edgeVectorT) {
	for i, ev := range evsToTruncate {
		otherEV := otherEVs[i]
		if ev.fromVertIdx != otherEV.fromVertIdx {
			log.Fatalf("truncateExtrusion: programming error: ev=%+v, otherEV=%+v", ev, otherEV)
		}
		for _, faceIdx := range is.edgeToFaces[ev.edge] {
			// log.Printf("truncateExtrusion: shortening edge %v on faceIdx=%v from vertIdx=%v to vertIdx=%v", ev.edge, faceIdx, ev.fromVertIdx, otherEV.toVertIdx)
			is.replaceFaceVertIdx(faceIdx, ev.fromVertIdx, otherEV.toVertIdx)
		}
	}
}
