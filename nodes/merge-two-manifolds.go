package nodes

import (
	"log"

	"github.com/gmlewis/advent-of-code-2021/enum"
)

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	// step 1 - find all shared vertices
	sharedVerts := map[VertIndexT]int{}
	bumpVertCount := func(vertIdx VertIndexT) { sharedVerts[vertIdx]++ }
	bumpVerts := func(f FaceT) { enum.Each(f, bumpVertCount) }
	enum.Each(fi.srcFaces, bumpVerts)
	enum.Each(fi.dstFaces, bumpVerts)

	log.Fatalf("merge2manifolds - not yet implemented")
	// last step: combine face sets
	fi.m.Faces = append(fi.dstFaces, fi.srcFaces...)
}

func (f *faceInfoT) findSharedVerts() map[int][]int {
	result := map[int][]int{}
	return result
}
