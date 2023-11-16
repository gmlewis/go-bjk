package nodes

import "log"

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh.
func (fi *faceInfoT) mergeNonManifoldSrc() {
	for edge, faceIdxes := range fi.src.badEdges {
		for _, faceIdx := range faceIdxes {
			log.Printf("mergeNonManifoldSrc: edge %v: bad face[%v] normal=%v", edge, faceIdx, fi.src.faceNormals[faceIdx])
		}
	}

	log.Printf("WARNING: mergeNonManifoldSrc - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dst.faces, fi.src.faces...)
}
