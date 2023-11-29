// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"slices"

	"github.com/gmlewis/advent-of-code-2021/enum"
)

func (fi *faceInfoT) checkForIntersectFaces() {
	log.Printf("\n\ncheckForIntersectFaces: %v src faces", len(fi.src.faces))
	log.Printf("checkForIntersectFaces: %v dst faces", len(fi.dst.faces))

	srcNormals := map[vertKeyT][]faceIndexT{}
	for i := range fi.src.faces {
		srcFaceIdx := faceIndexT(i)
		key := fi.src.faceNormals[i].toKey()
		srcNormals[key] = append(srcNormals[key], srcFaceIdx)
	}
	log.Printf("%v unique srcNormals", len(srcNormals))

	sharedNormals := map[vertKeyT][2][]faceIndexT{}
	for i := range fi.dst.faces {
		dstFaceIdx := faceIndexT(i)
		// we are interested in abutting faces with opposite normals, so negate the dst normals.
		key := fi.dst.faceNormals[i].Negated().toKey()
		sn, ok := srcNormals[key]
		if !ok {
			continue
		}
		if sns, ok := sharedNormals[key]; ok {
			sns[1] = append(sns[1], dstFaceIdx)
		} else {
			sharedNormals[key] = [2][]faceIndexT{sn, {dstFaceIdx}}
		}
	}
	log.Printf("%v shared normals", len(sharedNormals))

	for key, v := range sharedNormals {
		log.Printf("\n\nshared normal %v src (%v) faces:\n%v", key, len(v[0]), fi.src.dumpFacesByIndices(v[0]))
		log.Printf("shared normal %v dst (%v) faces:\n%v", key, len(v[1]), fi.dst.dumpFacesByIndices(v[1]))
		if len(v[0]) != 1 || len(v[1]) != 1 {
			log.Printf("WARNING: checkForIntersectFaces: unhandled case: #src=%v, #dst=%v, want 1,1", len(v[0]), len(v[1]))
			continue
		}

		// as a temporary hack, multiply all the vertices in each face by the sharedNormal and if they are
		// all equal, then chances are high that we have an intersected face.
		srcFaceIdx, dstFaceIdx := v[0][0], v[1][0]
		sharedNormal := fi.src.faceNormals[srcFaceIdx]
		var refResult Vec3
		multMatches := func(i int, vertIdx VertIndexT) bool {
			result := Vec3Mul(fi.m.Verts[vertIdx], sharedNormal)
			if i == 0 {
				refResult = result
				return true
			}
			return result.AboutEq(refResult)
		}

		if !enum.AllWithIndex(fi.src.faces[srcFaceIdx], multMatches) {
			continue
		}

		multInverseMatches := func(vertIdx VertIndexT) bool { return multMatches(1, vertIdx) }
		if !enum.All(fi.dst.faces[dstFaceIdx], multInverseMatches) {
			continue
		}

		fi.cutIntersectingFaces(srcFaceIdx, dstFaceIdx)
	}
}

func (fi *faceInfoT) cutIntersectingFaces(srcFaceIdx, dstFaceIdx faceIndexT) {
	log.Printf("cutting intersecting faces: srcFaceIdx=%v, dstFaceIdx=%v", srcFaceIdx, dstFaceIdx)
	srcFaceArea := fi.m.faceArea(fi.src.faces[srcFaceIdx])
	dstFaceArea := fi.m.faceArea(fi.dst.faces[dstFaceIdx])
	if dstFaceArea < srcFaceArea {
		log.Printf("WARNING: cutIntersectingFaces: unhandled dstFaceArea=%v < srcFaceArea=%v", dstFaceArea, srcFaceArea)
		return
	}

	fi.src.facesTargetedForDeletion[srcFaceIdx] = true

	srcFace, dstFace := fi.src.faces[srcFaceIdx], fi.dst.faces[dstFaceIdx]
	slices.Reverse(srcFace)
	log.Printf("reversed src face:\n%v", fi.m.dumpFace(srcFaceIdx, srcFace))

	// pick two random vertices from dstFace at opposite edges
	dstI0, dstI1 := 0, len(dstFace)/2
	dstVertIdx0, dstVertIdx1 := dstFace[dstI0], dstFace[dstI1]
	srcI0, srcVertIdx0 := fi.closestVertOnFace(dstVertIdx0, srcFace)
	srcI1, srcVertIdx1 := fi.closestVertOnFace(dstVertIdx1, srcFace)
	log.Printf("making edge from dstVertIdx0=%v(@%v) to srcVertIdx0=%v(@%v)", dstVertIdx0, dstI0, srcVertIdx0, srcI0)
	log.Printf("making edge from dstVertIdx1=%v(@%v) to srcVertIdx1=%v(@%v)", dstVertIdx1, dstI1, srcVertIdx1, srcI1)

	newDstFace0 := FaceT{dstVertIdx0}
	for i := range srcFace {
		vIdx := srcFace[(i+srcI0)%len(srcFace)]
		newDstFace0 = append(newDstFace0, vIdx)
		if vIdx == srcVertIdx1 {
			break
		}
	}
	newDstFace0 = append(newDstFace0, dstFace[1:dstI1+1]...)
	log.Printf("newDstFace0=%+v", newDstFace0)
	fi.dst.faces[dstFaceIdx] = newDstFace0

	newDstFace1 := FaceT{dstVertIdx0, dstVertIdx1}
	for i := range srcFace {
		vIdx := srcFace[(i+srcI1)%len(srcFace)]
		newDstFace1 = append(newDstFace1, vIdx)
		if vIdx == srcVertIdx0 {
			break
		}
	}
	newDstFace1 = append(newDstFace1, dstFace[dstI1+1:]...)
	log.Printf("newDstFace1=%+v", newDstFace1)
	fi.dst.faces = append(fi.dst.faces, newDstFace1)
}

/*
face[0]={[870 881 880 879 878 877 876 875 874 873 872 871]}: {{14.81 49.50 1.95} {14.88 49.50 2.20} {14.81 49.50 2.45} {14.63 49.50 2.63} {14.38 49.50 2.70} {14.13 49.50 2.63} {13.95 49.50 2.45} {13.88 49.50 2.20} {13.95 49.50 1.95} {14.13 49.50 1.76} {14.38 49.50 1.70} {14.63 49.50 1.76}}
2023/11/28 22:08:59 shared normal 0.000 -1.000 0.000 dst (1) faces:
face[6]={[6 7 8 9 10 11]}: {{18.00 49.50 0.00} {17.82 49.50 2.57} {17.27 49.50 5.08} {18.31 49.50 5.42} {18.90 49.50 2.74} {19.10 49.50 0.00}}
2023/11/28 22:35:47 cutting intersecting faces: srcFaceIdx=0, dstFaceIdx=6
2023/11/28 22:35:47 reversed src face:
face[0]={[871 872 873 874 875 876 877 878 879 880 881 870]}: {{14.63 49.50 1.76} {14.38 49.50 1.70} {14.13 49.50 1.76} {13.95 49.50 1.95} {13.88 49.50 2.20} {13.95 49.50 2.45} {14.13 49.50 2.63} {14.38 49.50 2.70} {14.63 49.50 2.63} {14.81 49.50 2.45} {14.88 49.50 2.20} {14.81 49.50 1.95}}
2023/11/28 22:35:47 making edge from dstVertIdx0=6(@0) to srcVertIdx0=870(@11)
2023/11/28 22:35:47 making edge from dstVertIdx1=9(@3) to srcVertIdx1=880(@9)
2023/11/28 22:46:20 newDstFace0=[6 870 871 872 873 874 875 876 877 878 879 880 7 8 9]
2023/11/28 22:48:24 newDstFace1=[6 9 880 881 870 10 11]

6 870 871 872 873 874 875 876 877 878 879 880 7 8 9
9 880 881 870 10 11 6
*/
