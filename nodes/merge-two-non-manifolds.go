package nodes

import "log"

// merge2NonManifolds merges two non-manifold srcFaces and dstFaces meshes together,
// but may or may not create a resulting manifold mesh due to the possibility of one of
// the meshes still being un-capped (such as the case of the helix). However, the resulting
// mesh should be manifold apart from any unconnected parts of the original meshes.
// In other words, there should only remain edges that have exactly 1 or 2 faces, not more.
func (fi *faceInfoT) merge2NonManifolds() {
	log.Fatalf("merge2NonManifolds - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dst.faces, fi.src.faces...)
}
