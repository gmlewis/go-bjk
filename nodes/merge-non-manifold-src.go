package nodes

import (
	"fmt"
	"log"
)

// mergeNonManifoldSrc merges the non-manifold srcFaces mesh into the manifold dstFaces mesh,
// creating a final manifold mesh (ideally, although it is possible that it is still non-manifold).
func (fi *faceInfoT) mergeNonManifoldSrc() {
	// first step - if all bad edges are owned by a set of faces with two edges each,
	// chances are high that those faces should simply be deleted.
	srcFaceIndicesToEdges := reverseMapBadEdges(fi.src.badEdges)
	debugFaces := make([]FaceT, 0, len(srcFaceIndicesToEdges))
	log.Printf("mergeNonManifoldSrc: srcFaceIndicesToEdges: %+v", srcFaceIndicesToEdges)
	for srcFaceIdx, badEdges := range srcFaceIndicesToEdges {
		debugFaces = append(debugFaces, fi.src.faces[srcFaceIdx])
		log.Printf("mergeNonManifoldSrc: src.faces[%v] has %v bad edges: %+v", srcFaceIdx, len(badEdges), badEdges)
	}
	fi.m.Faces = debugFaces
	fi.m.WriteSTL(fmt.Sprintf("debug-%v-%v-badFaces-%v.stl", len(fi.src.faces), len(fi.dst.faces), len(debugFaces)))

	for edge, faceIdxes := range fi.src.badEdges {
		switch len(faceIdxes) {
		case 1:
			log.Printf("mergeNonManifoldSrc: skipping edge %v with one face", edge)
		case 3:
			debugFileBaseName := fmt.Sprintf("debug-%v-%v-edge-%v-%v-", len(fi.src.faces), len(fi.dst.faces), edge[0], edge[1])
			log.Printf("debugFileBaseName=%v", debugFileBaseName)
			fi.m.Faces = fi.src.faces
			fi.m.WriteSTL(debugFileBaseName + "src.stl")
			fi.m.Faces = fi.dst.faces
			fi.m.WriteSTL(debugFileBaseName + "dst.stl")

			fi.src.fixEdge3Faces(edge, faceIdxes, debugFileBaseName)
		default:
			log.Printf("WARNING: mergeNonManifoldSrc: skipping edge %v with %v faces", edge, len(faceIdxes))
		}
	}
}

func (is *infoSetT) fixEdge3Faces(edge edgeT, faceIdxes []faceIndexT, debugFileBaseName string) {
	f0, f1, f2 := faceIdxes[0], faceIdxes[1], faceIdxes[2]
	switch {
	case is.faceNormals[f0].AboutEq(is.faceNormals[f1]):
		is.fixEdge2OverlapingFaces(edge, f0, f1, f2, debugFileBaseName)
	case is.faceNormals[f1].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f1, f2, f0, debugFileBaseName)
	case is.faceNormals[f0].AboutEq(is.faceNormals[f2]):
		is.fixEdge2OverlapingFaces(edge, f0, f2, f1, debugFileBaseName)
	default:
		log.Printf("WARNING: fixEdge3Faces: unhandled case normals: %v %v %v", is.faceNormals[f0], is.faceNormals[f1], is.faceNormals[f2])
	}
}

func (is *infoSetT) fixEdge2OverlapingFaces(edge edgeT, f0, f1, otherFaceIdx faceIndexT, debugFileBaseName string) {
	log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f0=%v: %v", edge, f0, is.faceNormals[f0], is.faceInfo.m.dumpFace(f0, is.faces[f0]))
	log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, f1=%v: %v", edge, f1, is.faceNormals[f1], is.faceInfo.m.dumpFace(f1, is.faces[f1]))
	log.Printf("fixEdge2OverlapingFaces(edge=%v): normal=%v, otherFace=%v: %v", edge, otherFaceIdx, is.faceNormals[otherFaceIdx], is.faceInfo.m.dumpFace(otherFaceIdx, is.faces[otherFaceIdx]))
	log.Printf("fixEdge2OverlapingFaces(edge=%v): shared edge: %v - %v", edge, is.faceInfo.m.Verts[edge[0]], is.faceInfo.m.Verts[edge[1]])

	is.faceInfo.m.Faces = []FaceT{is.faces[f0]}
	is.faceInfo.m.WriteSTL(debugFileBaseName + "f0.stl")
	is.faceInfo.m.Faces = []FaceT{is.faces[f1]}
	is.faceInfo.m.WriteSTL(debugFileBaseName + "f1.stl")
	is.faceInfo.m.Faces = []FaceT{is.faces[otherFaceIdx]}
	is.faceInfo.m.WriteSTL(debugFileBaseName + "otherFaceIdx.stl")

	log.Fatalf("STOP")

	// is.facesTargetedForDeletion[otherFaceIdx] = true

	// f0VertKey := is.toVertKey(is.faces[f0])
	// f1VertKey := is.toVertKey(is.faces[f1])
	// if f0VertKey == f1VertKey {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces are identical! f0VertKey=%v", edge, f0VertKey)
	// } else {
	// 	log.Printf("fixEdge2OverlapingFaces(edge=%v): faces DIFFER!\nf0VertKey=%v\nf1VertKey=%v", edge, f0VertKey, f1VertKey)
	// }
}
