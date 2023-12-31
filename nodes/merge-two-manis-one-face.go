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

	// all edgeVectors for src are identical and all EVs for dst are also identical.
	// Find out which is shorter and delete the whole thing.
	if srcSideEVs[0].length > dstSideEVs[0].length {
		// If all the verts of the shorter (dst) side are only used by this dstOtherEndFace, all
		// the faces of this dst object can be deleted, leaving only the src object!
		dstFaceToDeleteIdx, ok := fi.dst.faceStrToFaceIdx[dstOtherEndFace.toKey()]
		if !ok {
			log.Fatalf("mergeExtrusion: unable to get dstFace to delete from %+v", dstOtherEndFace)
		}
		fi.dst.facesTargetedForDeletion[dstFaceToDeleteIdx] = true
		fi.dst.facesTargetedForDeletion[dstFaceIdx] = true
		fi.dst.deleteSideFaces(dstSideEVs)

		return
	}

	// If all the verts of the shorter (src) side are only used by this srcOtherEndFace, all
	// the faces of this src object can be deleted, leaving only the dst object!
	srcFaceToDeleteIdx, ok := fi.src.faceStrToFaceIdx[srcOtherEndFace.toKey()]
	if !ok {
		log.Fatalf("mergeExtrusion: unable to get srcFace to delete from %+v", srcOtherEndFace)
	}
	fi.src.facesTargetedForDeletion[srcFaceToDeleteIdx] = true
	fi.src.facesTargetedForDeletion[srcFaceIdx] = true
	fi.src.deleteSideFaces(srcSideEVs)
}

func (is *infoSetT) deleteSideFaces(evs []edgeVectorT) {
	for _, ev := range evs {
		for _, faceIdx := range is.edgeToFaces[ev.edge] {
			// log.Printf("deleteSideFaces: deleting faceIdx=%v", faceIdx)
			is.facesTargetedForDeletion[faceIdx] = true
		}
	}
}
