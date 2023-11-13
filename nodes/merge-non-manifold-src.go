package nodes

import "log"

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh.
func (fi *faceInfoT) mergeNonManifoldSrc() {
	log.Fatalf("mergeNonManifoldSrc - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}
