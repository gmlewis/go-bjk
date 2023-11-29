// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"

	"golang.org/x/exp/maps"
)

// merge2manifolds merges the manifold srcFaces and dstFaces meshes together,
// creating a final manifold mesh.
func (fi *faceInfoT) merge2manifolds() {
	// step 1 - find all shared vertices, edges, and faces
	sharedVerts, sharedEdges, sharedFaces := fi.findSharedVEFs()
	// log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)
	// log.Printf("merge2manifolds: shared edges: %+v", sharedEdges)
	// log.Printf("merge2manifolds: shared faces: %+v", sharedFaces)

	switch {
	case len(sharedFaces) > 1:
		log.Printf("WARNING: merge2manifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedFaces) == 1:
		key := maps.Keys(sharedFaces)[0]
		fi.merge2manisOneFace(sharedEdges, sharedFaces[key][0], sharedFaces[key][1])
	case len(sharedEdges) == 2:
		fi.merge2manis2edges(sharedEdges)
	case len(sharedEdges) > 1:
		fi.merge2manisManyEdges(sharedEdges)
	case len(sharedEdges) == 1:
		fi.merge2manis2edges(sharedEdges) // experiment
	case len(sharedVerts) == 0 && len(sharedEdges) == 0 && len(sharedFaces) == 0: // simple concatenation - no sharing
		fi.checkForIntersectFaces()
	default:
		log.Printf("WARNING: merge2manifolds - unhandled shares: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}
}

func (fi *faceInfoT) merge2manis2edges(sharedEdges sharedEdgesMapT) {
	srcFaceIndicesToEdges, dstFaceIndicesToEdges := reverseMapFaceIndicesToEdges(sharedEdges)
	if len(srcFaceIndicesToEdges) > len(dstFaceIndicesToEdges) {
		fi.swapSrcAndDst(nil)
		for edge, v := range sharedEdges {
			sharedEdges[edge] = [2][]faceIndexT{v[1], v[0]}
		}
		srcFaceIndicesToEdges, dstFaceIndicesToEdges = dstFaceIndicesToEdges, srcFaceIndicesToEdges
	}
	fi.merge2manis2edgesSrcFewerThanDst(sharedEdges, srcFaceIndicesToEdges, dstFaceIndicesToEdges)
}
