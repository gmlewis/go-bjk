package nodes

import (
	"fmt"
	"log"
)

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
	uniqueVertsMap := map[vertKeyT]VertIndexT{}
	vertsOldToNew := make([]VertIndexT, 0, len(verts))
	uniqueVerts := make([]Vec3, 0, len(verts)) // this estimate is too large, but it is order-of-ballpark correct.
	for _, vert := range verts {
		s := vert.toKey()
		if idx, ok := uniqueVertsMap[s]; ok {
			vertsOldToNew = append(vertsOldToNew, idx)
			continue
		}
		newIdx := VertIndexT(len(uniqueVerts))
		vertsOldToNew = append(vertsOldToNew, newIdx)
		uniqueVertsMap[s] = newIdx
		uniqueVerts = append(uniqueVerts, vert)
	}
	dst.Verts = uniqueVerts
	dst.uniqueVerts = uniqueVertsMap

	adjFace := func(face FaceT, offset int) FaceT {
		result := make(FaceT, 0, len(face))
		for _, vIdx := range face {
			result = append(result, vertsOldToNew[int(vIdx)+offset])
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
	// dst.Faces = append(faces, srcFaces...) // ONLY FOR DEBUGGING WHEN NOT RUNNING MANIFOLD MERGE!!!

	// log.Printf("\n\nAFTER MERGE:\nfaces:\n%v", dst.dumpFaces(dst.Faces))

	// verify that this step did not create non-manifold geometry.
	fi := dst.genFaceInfo(dst.Faces, nil)
	if len(fi.dst.badEdges) > 0 {
		fi.m.WriteSTL(fmt.Sprintf("after-merge-badDstEdges-%v-src.stl", len(fi.dst.badEdges)))

		for edge, faceIdxes := range fi.dst.badEdges {
			for _, faceIdx := range faceIdxes {
				log.Printf("NEW BAD EDGE: %v: %v", edge, dst.dumpFace(faceIdx, dst.Faces[faceIdx]))
			}
		}

		log.Fatalf("Merg: BAD MERGE STOP")
	}
}

func (dst *Mesh) manifoldMerge(dstFaces, srcFaces []FaceT) {
	// log.Printf("\n\nmanifoldMerge: srcFaces=%+v\n%v", srcFaces, dst.dumpFaces(srcFaces))
	// log.Printf("manifoldMerge: dstFaces=%+v\n%v", dstFaces, dst.dumpFaces(dstFaces))

	fi := dst.genFaceInfo(dstFaces, srcFaces)
	// log.Printf("manifoldMerge: src.badEdges=%v=%+v", len(fi.src.badEdges), fi.src.badEdges)
	// log.Printf("manifoldMerge: dst.badEdges=%v=%+v", len(fi.dst.badEdges), fi.dst.badEdges)
	// fi.m.Faces = fi.src.faces
	// fi.m.WriteSTL(fmt.Sprintf("before-merge-badSrcEdges-%v-badDstEdges-%v-src.stl", len(fi.src.badEdges), len(fi.dst.badEdges)))
	// fi.m.Faces = fi.dst.faces
	// fi.m.WriteSTL(fmt.Sprintf("before-merge-badSrcEdges-%v-badDstEdges-%v-dst.stl", len(fi.src.badEdges), len(fi.dst.badEdges)))

	switch {
	case len(fi.src.badEdges) == 0 && len(fi.dst.badEdges) == 0:
		fi.merge2manifolds()
	case len(fi.src.badEdges) == 0:
		// swap src and dst so that src is the non-manifold mesh:
		fi.swapSrcAndDst()
		fi.mergeNonManifoldSrc()
	case len(fi.dst.badEdges) == 0:
		fi.mergeNonManifoldSrc()
	default:
		fi.merge2NonManifolds()
	}

	// last step: delete targeted faces for deletion, then combine.
	fi.src.deleteFacesLastToFirst(fi.src.facesTargetedForDeletion)
	fi.dst.deleteFacesLastToFirst(fi.dst.facesTargetedForDeletion)
	fi.m.Faces = append(fi.dst.faces, fi.src.faces...)
}
