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
	log.Printf("merge2manifolds: shared verts: %+v", sharedVerts)
	log.Printf("merge2manifolds: shared edges: %+v", sharedEdges)
	log.Printf("merge2manifolds: shared faces: %+v", sharedFaces)

	switch {
	case len(sharedFaces) > 0:
		log.Fatalf("merge2manifolds - unhandled: shared faces: %+v", sharedFaces)
	case len(sharedEdges) == 1:
		edges := maps.Keys(sharedEdges)
		edge := edges[0]
		fi.merge2manisOneEdge(sharedVerts, edge, sharedEdges[edge][0], sharedEdges[edge][1])
	default:
		log.Fatalf("merge2manifolds - unhandled: #verts=%v, #edges=%v, #faces=%v", len(sharedVerts), len(sharedEdges), len(sharedFaces))
	}

	// last step: combine face sets
	fi.m.Faces = append(fi.dst.faces, fi.src.faces...)
}

func (fi *faceInfoT) merge2manisOneEdge(sharedVerts sharedVertsMapT, edge edgeT, srcFaces, dstFaces []faceIndexT) {
	assert(len(srcFaces) == 2 && len(dstFaces) == 2, "merge2manisOneEdge: want 2 srcFaces and 2 dstFaces")

	// sort srcFaces by area (descending - largest first)
	if srcFace0Area, srcFace1Area := fi.m.faceArea(fi.src.faces[srcFaces[0]]), fi.m.faceArea(fi.src.faces[srcFaces[1]]); srcFace0Area < srcFace1Area {
		srcFaces[0], srcFaces[1] = srcFaces[1], srcFaces[0]
	}

	log.Printf("merge2manisOneEdge: sorted srcFaces by area:\n%v", fi.m.dumpFaces([]FaceT{fi.src.faces[srcFaces[0]], fi.src.faces[srcFaces[1]]}))
}
