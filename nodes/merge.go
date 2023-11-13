package nodes

import "log"

// Merge merges src into dst for Ops.merge(dst, src).
func (dst *Mesh) Merge(src *Mesh) {
	// First, a naive merge is performed by not checking if any Verts are shared.
	verts := make([]Vec3, 0, len(dst.Verts)+len(src.Verts))
	verts = append(verts, dst.Verts...)
	verts = append(verts, src.Verts...)

	// If there are no faces, then simply concatenate the verts/normals/tangents and return.
	if len(dst.Faces) == 0 && len(src.Faces) == 0 {
		normals := make([]Vec3, 0, len(dst.Normals)+len(src.Normals))
		tangents := make([]Vec3, 0, len(dst.Tangents)+len(src.Tangents))

		normals = append(normals, dst.Normals...)
		normals = append(normals, src.Normals...)
		tangents = append(tangents, dst.Tangents...)
		tangents = append(tangents, src.Tangents...)

		dst.Verts = verts
		dst.Normals = normals
		dst.Tangents = tangents
		return
	}

	// However, if there are faces, the normals and tangents are no longer usable; delete them.
	numOrigDstVerts := len(dst.Verts)
	dst.Normals = nil
	dst.Tangents = nil

	// Next, a map is made of unique verts with a mapping of old indices to new ones.
	uniqueVertsMap := map[string]int{}
	vertsOldToNew := make([]int, 0, len(verts))
	uniqueVerts := make([]Vec3, 0, len(verts)) // this estimate is too large, but it is order-of-ballpark correct.
	for _, vert := range verts {
		s := vert.String()
		if idx, ok := uniqueVertsMap[s]; ok {
			vertsOldToNew = append(vertsOldToNew, idx)
			continue
		}
		newIdx := len(uniqueVerts)
		vertsOldToNew = append(vertsOldToNew, newIdx)
		uniqueVertsMap[s] = newIdx
		uniqueVerts = append(uniqueVerts, vert)
	}
	dst.Verts = uniqueVerts
	// if len(verts) != len(uniqueVerts) {
	// 	log.Printf("Merge: reduced verts from %v to %v", len(verts), len(uniqueVerts))
	// }

	adjFace := func(face FaceT, offset int) FaceT {
		result := make(FaceT, 0, len(face))
		for _, vIdx := range face {
			result = append(result, vertsOldToNew[vIdx+offset])
		}
		return result
	}

	// Eventually, the src.Faces and dst.Faces will be merged.
	// However, keep them separate here to simplify the manifold merge algorithm
	faces := make([]FaceT, 0, len(dst.Faces)+len(src.Faces))
	srcFaces := make([]FaceT, 0, len(src.Faces))
	for _, face := range dst.Faces {
		faces = append(faces, adjFace(face, 0))
	}
	for _, face := range src.Faces {
		srcFaces = append(srcFaces, adjFace(face, numOrigDstVerts))
	}

	// Now, make sure that all faces will be manifold before combining.
	dst.manifoldMerge(faces, srcFaces)

	log.Fatalf("GML: DEBUGGING: First manifoldMerge")
}

func (dst *Mesh) manifoldMerge(dstFaces, srcFaces []FaceT) {
	log.Printf("\n\nmanifoldMerge: srcFaces=%+v\n%v", srcFaces, dst.dumpFaces(srcFaces))
	log.Printf("manifoldMerge: dstFaces=%+v\n%v", dstFaces, dst.dumpFaces(dstFaces))

	fi := dst.genFaceInfo(dstFaces, srcFaces)
	switch {
	case len(fi.srcBadEdges) == 0 && len(fi.dstBadEdges) == 0:
		fi.merge2manifolds()
	case len(fi.srcBadEdges) == 0:
		// swap src and dst so that src is the non-manifold mesh:
		fi.swapSrcAndDst()
		fi.mergeNonManifoldSrc()
	case len(fi.dstBadEdges) == 0:
		fi.mergeNonManifoldSrc()
	default:
		fi.merge2NonManifolds()
	}
}

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	log.Fatalf("merge2manifolds - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh.
func (fi *faceInfoT) mergeNonManifoldSrc() {
	log.Fatalf("mergeNonManifoldSrc - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}

// merge2NonManifolds merges two non-manifold srcFaces and dstFaces meshes together,
// but may or may not create a resulting manifold mesh due to the possibility of one of
// the meshes still being un-capped (such as the case of the helix). However, the resulting
// mesh should be manifold apart from any unconnected parts of the original meshes.
// In other words, there should only remain edges that have exactly 1 or 2 faces, not more.
func (fi *faceInfoT) merge2NonManifolds() {
	log.Fatalf("merge2NonManifolds - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}
