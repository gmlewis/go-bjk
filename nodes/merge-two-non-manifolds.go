package nodes

import (
	"log"

	"golang.org/x/exp/maps"
)

// merge2NonManifolds merges two non-manifold srcFaces and dstFaces meshes together,
// but may or may not create a resulting manifold mesh due to the possibility of one of
// the meshes still being un-capped (such as the case of the helix). However, the resulting
// mesh should be manifold apart from any unconnected parts of the original meshes.
// In other words, there should only remain edges that have exactly 1 or 2 faces, not more.
func (fi *faceInfoT) merge2NonManifolds() {
	// step 1 - find all shared vertices, edges, and faces
	sharedVerts, sharedEdges, sharedFaces := fi.findSharedVEFs()
	log.Printf("merge2NonManifolds: shared verts: %+v", sharedVerts)
	log.Printf("merge2NonManifolds: shared edges: %+v", sharedEdges)
	log.Printf("merge2NonManifolds: shared faces: %+v", sharedFaces)

	switch {
	case len(sharedFaces) > 0:
		log.Printf("WARNING: merge2NonManifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedEdges) > 1:
		fi.merge2NonManisManyEdges(sharedEdges)
	case len(sharedEdges) == 1:
		edges := maps.Keys(sharedEdges)
		edge := edges[0]
		fi.merge2NonManisOneEdge(sharedVerts, edge, sharedEdges[edge][0], sharedEdges[edge][1])
	case len(sharedVerts) == 0 && len(sharedEdges) == 0 && len(sharedFaces) == 0: // simple concatenation - no sharing
	default:
		log.Printf("WARNING: merge2NonManifolds - unhandled shares: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}
}

func (fi *faceInfoT) merge2NonManisOneEdge(sharedVerts sharedVertsMapT, edge edgeT, srcFaces, dstFaces []faceIndexT) {
	assert(len(srcFaces) == 2 && len(dstFaces) == 2, "merge2NonManisOneEdge: want 2 srcFaces and 2 dstFaces")

	log.Printf("WARNING: merge2NonManisOneEdge - not implemented yet")
}
