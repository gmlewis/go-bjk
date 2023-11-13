package nodes

import (
	"log"
)

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	// step 1 - find all shared vertices
	sharedVerts := fi.findSharedVerts()
	log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)

	log.Fatalf("merge2manifolds - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}
