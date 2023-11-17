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

/*
2023/11/16 20:47:24 manifoldMerge: src.badEdges=0=map[]
2023/11/16 20:47:24 manifoldMerge: dst.badEdges=0=map[]

2023/11/16 20:47:24 merge2manifolds: shared verts: map[70:[[0 1 9] [33 46 47]] 71:[[0 1 2] [33 34 47]] 72:[[0 2 3] [34 35 47]] 73:[[0 3 4] [35 36 47]] 74:[[0 4 5] [36 37 47]] 75:[[0 5 6] [37 38 47]] 76:[[0 6 7] [38 39 47]] 236:[[7 8 10] [145 146 149 150]] 237:[[8 9 10] [146 147 150 151]] 240:[[0 7 8] [149 150 152]] 241:[[0 8 9] [150 151 152]]]

2023/11/16 20:47:24 merge2manifolds: shared edges: map[[70 71]:[[0 1] [33 47]] [71 72]:[[0 2] [34 47]] [72 73]:[[0 3] [35 47]] [73 74]:[[0 4] [36 47]] [74 75]:[[0 5] [37 47]] [75 76]:[[0 6] [38 47]] [236 237]:[[8 10] [146 150]] [236 240]:[[7 8] [149 150]] [237 241]:[[8 9] [150 151]] [240 241]:[[0 8] [150 152]]]

2023/11/16 20:47:24 merge2manifolds: shared faces: map[[236 237 240 241]:[8 150]]

srcFaces:...
face[8]={[236 240 241 237]}: {{-6.50 6.50 0.00} {-6.50 7.50 0.00} {-6.50 7.50 1.00} {-6.50 6.50 1.00}}

dstFaces:...
face[150]={[240 236 237 241]}: {{-6.50 7.50 0.00} {-6.50 6.50 0.00} {-6.50 6.50 1.00} {-6.50 7.50 1.00}}

2023/11/16 20:47:24 WARNING: merge2manisOneFace unhandled case:
srcFaceNumVerts=4,
dstFaceNumVerts=4,
sharedEdges=10=map[
[70 71]:[[0 1] [33 47]]
[71 72]:[[0 2] [34 47]]
[72 73]:[[0 3] [35 47]]
[73 74]:[[0 4] [36 47]]
[74 75]:[[0 5] [37 47]]
[75 76]:[[0 6] [38 47]]
[236 237]:[[8 10] [146 150]]
[236 240]:[[7 8] [149 150]]
[237 241]:[[8 9] [150 151]]
[240 241]:[[0 8] [150 152]]]
*/

func (fi *faceInfoT) mergeExtrusion(sharedEdges sharedEdgesMapT, srcFaceIdx, dstFaceIdx faceIndexT) {
	log.Printf("mergeExtrusion: sharedEdges=%+v", sharedEdges)
	log.Printf("mergeExtrusion: srcFaceIdx=%v, dstFaceIdx=%v", srcFaceIdx, dstFaceIdx)

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

	log.Fatalf("mergeExtrusion: STOP")
}

func (fi *faceInfoT) truncateExtrusion(is *infoSetT, evsToTruncate, otherEVs []edgeVectorT) {
	for i, ev := range evsToTruncate {
		otherEV := otherEVs[i]
		if ev.fromVertIdx != otherEV.fromVertIdx {
			log.Fatalf("truncateExtrusion: programming error: ev=%+v, otherEV=%+v", ev, otherEV)
		}
		for _, faceIdx := range is.edgeToFaces[ev.edge] {
			log.Printf("truncateExtrusion: shortening edge %v on faceIdx=%v from vertIdx=%v to vertIdx=%v", ev.edge, faceIdx, ev.fromVertIdx, otherEV.toVertIdx)
			is.replaceFaceVertIdx(faceIdx, ev.fromVertIdx, otherEV.toVertIdx)
		}
	}
}

/*
2023/11/16 22:11:07 manifoldMerge: src.badEdges=0=map[]
2023/11/16 22:11:07 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 22:11:07 merge2manifolds: shared verts: map[379:[[0 3 4] [261 262 263]] 380:[[0 2 3] [261 263 275 277]] 383:[[3 4 5] [262 263 265]] 384:[[2 3 5] [263 265 274 275]]]
2023/11/16 22:11:07 merge2manifolds: shared edges: map[[379 380]:[[0 3] [261 263]] [379 383]:[[3 4] [262 263]] [380 384]:[[2 3] [263 275]] [383 384]:[[3 5] [263 265]]]
2023/11/16 22:11:07 merge2manifolds: shared faces: map[[379 380 383 384]:[3 263]]
2023/11/16 22:11:07 mergeExtrusion: sharedEdges=map[[379 380]:[[0 3] [261 263]] [379 383]:[[3 4] [262 263]] [380 384]:[[2 3] [263 275]] [383 384]:[[3 5] [263 265]]]
2023/11/16 22:11:07 mergeExtrusion: srcFaceIdx=3, dstFaceIdx=263
2023/11/16 22:11:07 connectedEdgeVectorFromVertOnFace(vertIdx=384, edge=[380 384], faceIdx=2): i=3, pIdx=384, nextIdx=403, returning ({-5.25 6.50 -9.09}).Sub({-8.99 6.50 -5.43})
2023/11/16 22:11:07 connectedEdgeVectorFromVertOnFace(vertIdx=380, edge=[379 380], faceIdx=0): i=1, pIdx=380, nextIdx=401, returning ({-5.25 5.50 -9.09}).Sub({-8.99 5.50 -5.43})
2023/11/16 22:11:07 connectedEdgeVectorFromVertOnFace(vertIdx=379, edge=[379 383], faceIdx=4): i=1, pIdx=379, nextIdx=400, returning ({-4.75 5.50 -8.23}).Sub({-8.13 5.50 -4.91})
2023/11/16 22:11:07 connectedEdgeVectorFromVertOnFace(vertIdx=383, edge=[383 384], faceIdx=5): i=3, pIdx=383, nextIdx=402, returning ({-4.75 6.50 -8.23}).Sub({-8.13 6.50 -4.91})

2023/11/16 22:11:07 mergeExtrusion: unhandled case: src
    ev={edge:[380 401] fromVertIdx:380 toVertIdx:401 toSubFrom:{X:3.736426161583494 Y:0 Z:-3.6624155318311846} length:5.232032873440673}
nextEV={edge:[379 400] fromVertIdx:379 toVertIdx:400 toSubFrom:{X:3.3805760509564937 Y:0 Z:-3.313614052609166} length:4.733744028351084}
*/

/*
2023/11/16 19:29:24 manifoldMerge: src.badEdges=0=map[]
2023/11/16 19:29:24 manifoldMerge: dst.badEdges=0=map[]
2023/11/16 19:29:24 merge2manifolds: shared verts: map[230:[[0 1 4] [143 144 147]] 231:[[0 1 2] [143 144 145]] 232:[[0 2 3] [143 145 146]] 233:[[0 3 4] [143 146 147]]]
2023/11/16 19:29:24 merge2manifolds: shared edges: map[[230 231]:[[0 1] [143 144]] [230 233]:[[0 4] [143 147]] [231 232]:[[0 2] [143 145]] [232 233]:[[0 3] [143 146]]]
2023/11/16 19:29:24 merge2manifolds: shared faces: map[[230 231 232 233]:[0 143]]
2023/11/16 19:29:24 mergeExtrusion: sharedEdges=map[[230 231]:[[0 1] [143 144]] [230 233]:[[0 4] [143 147]] [231 232]:[[0 2] [143 145]] [232 233]:[[0 3] [143 146]]]
2023/11/16 19:29:24 mergeExtrusion: srcFaceIdx=0, dstFaceIdx=143

2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=233, edge=[232 233], faceIdx=3): i=2, pIdx=233, nextIdx=241, returning ({-6.50 7.50 1.00}).Sub({-6.50 3.50 1.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=232, edge=[231 232], faceIdx=2): i=2, pIdx=232, nextIdx=240, returning ({-6.50 7.50 0.00}).Sub({-6.50 3.50 0.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=231, edge=[230 231], faceIdx=1): i=2, pIdx=231, nextIdx=239, returning ({-5.50 7.50 0.00}).Sub({-5.50 3.50 0.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=230, edge=[230 233], faceIdx=4): i=2, pIdx=230, nextIdx=238, returning ({-5.50 7.50 1.00}).Sub({-5.50 3.50 1.00})

2023/11/16 19:29:24 srcSideEV: {edge:[233 241] fromVertIdx:233 toVertIdx:241 toSubFrom:{X:0 Y:4 Z:0} length:4}
2023/11/16 19:29:24 srcSideEV: {edge:[232 240] fromVertIdx:232 toVertIdx:240 toSubFrom:{X:0 Y:4 Z:0} length:4}
2023/11/16 19:29:24 srcSideEV: {edge:[231 239] fromVertIdx:231 toVertIdx:239 toSubFrom:{X:0 Y:4 Z:0} length:4}
2023/11/16 19:29:24 srcSideEV: {edge:[230 238] fromVertIdx:230 toVertIdx:238 toSubFrom:{X:0 Y:4 Z:0} length:4}

2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=233, edge=[232 233], faceIdx=146): i=2, pIdx=233, nextIdx=237, returning ({-6.50 6.50 1.00}).Sub({-6.50 3.50 1.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=232, edge=[231 232], faceIdx=145): i=2, pIdx=232, nextIdx=236, returning ({-6.50 6.50 0.00}).Sub({-6.50 3.50 0.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=231, edge=[230 231], faceIdx=144): i=2, pIdx=231, nextIdx=235, returning ({-5.50 6.50 0.00}).Sub({-5.50 3.50 0.00})
2023/11/16 19:29:24 connectedEdgeVectorFromVertOnFace(vertIdx=230, edge=[230 233], faceIdx=147): i=2, pIdx=230, nextIdx=234, returning ({-5.50 6.50 1.00}).Sub({-5.50 3.50 1.00})

2023/11/16 19:29:24 dstSideEV: {edge:[233 237] fromVertIdx:233 toVertIdx:237 toSubFrom:{X:0 Y:3 Z:0} length:3}
2023/11/16 19:29:24 dstSideEV: {edge:[232 236] fromVertIdx:232 toVertIdx:236 toSubFrom:{X:0 Y:3 Z:0} length:3}
2023/11/16 19:29:24 dstSideEV: {edge:[231 235] fromVertIdx:231 toVertIdx:235 toSubFrom:{X:0 Y:3 Z:0} length:3}
2023/11/16 19:29:24 dstSideEV: {edge:[230 234] fromVertIdx:230 toVertIdx:234 toSubFrom:{X:0 Y:3 Z:0} length:3}
*/
