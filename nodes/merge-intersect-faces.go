// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"math"

	"github.com/gmlewis/advent-of-code-2021/enum"
)

func (fi *faceInfoT) checkForIntersectFaces() {
	// log.Printf("\n\ncheckForIntersectFaces: %v src faces", len(fi.src.faces))
	// log.Printf("checkForIntersectFaces: %v dst faces", len(fi.dst.faces))

	srcNormals := map[vertKeyT][]faceIndexT{}
	for i := range fi.src.faces {
		srcFaceIdx := faceIndexT(i)
		key := fi.src.faceNormals[i].toKey()
		srcNormals[key] = append(srcNormals[key], srcFaceIdx)
		// log.Printf("srcFaceIdx=%v normal key {%v} #%v", srcFaceIdx, key, len(srcNormals[key]))
	}
	// log.Printf("%v unique srcNormals", len(srcNormals))

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
			sharedNormals[key] = sns
			// log.Printf("dstFaceIdx=%v normal key {%v} #%v", dstFaceIdx, key, len(sharedNormals[key][1]))
		} else {
			sharedNormals[key] = [2][]faceIndexT{sn, {dstFaceIdx}}
			// log.Printf("dstFaceIdx=%v normal key {%v} #%v", dstFaceIdx, key, len(sharedNormals[key][1]))
		}
	}
	// log.Printf("%v shared normals", len(sharedNormals))

	for _, v := range sharedNormals {
		for _, srcFaceIdx := range v[0] {
			srcFace := fi.src.faces[srcFaceIdx]
			dstFaceIdx := v[1][0]
			if len(v[1]) != 1 {
				dstFaceIdx = fi.findClosestFaceByCentroid(srcFace, v[1])
			}

			// as a temporary hack, multiply all the vertices in each face by the sharedNormal and if they are
			// all equal (and not zero), then chances are high that we have an intersected face.
			sharedNormal := fi.src.faceNormals[srcFaceIdx]
			var refResult Vec3
			multMatches := func(i int, vertIdx VertIndexT) bool {
				result := Vec3Mul(fi.m.Verts[vertIdx], sharedNormal)
				if i == 0 {
					refResult = result
					return !refResult.AboutZero()
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

			// log.Printf("\n\nshared normal %v:\nsrc face: %v", key, fi.m.dumpFace(srcFaceIdx, srcFace))
			// log.Printf("shared normal %v:\ndst face: %v", key, fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))

			fi.cutIntersectingFaces(srcFaceIdx, dstFaceIdx)
		}
	}
}

func (fi *faceInfoT) cutIntersectingFaces(srcFaceIdx, dstFaceIdx faceIndexT) {
	// log.Printf("cutting intersecting faces: srcFaceIdx=%v, dstFaceIdx=%v", srcFaceIdx, dstFaceIdx)
	srcFaceArea := fi.m.faceArea(fi.src.faces[srcFaceIdx])
	dstFaceArea := fi.m.faceArea(fi.dst.faces[dstFaceIdx])
	if dstFaceArea < srcFaceArea {
		log.Printf("WARNING: cutIntersectingFaces: unhandled dstFaceArea=%v < srcFaceArea=%v", dstFaceArea, srcFaceArea)
		return
	}

	fi.src.facesTargetedForDeletion[srcFaceIdx] = true

	srcFace, dstFace := fi.src.faces[srcFaceIdx], fi.dst.faces[dstFaceIdx]

	dstIToSrcI := map[int]int{}
	for i, dstVertIdx := range dstFace {
		dstIToSrcI[i] = fi.closestVertIOnFace(dstVertIdx, srcFace)
	}
	// log.Printf("dstIToSrcI=%+v", dstIToSrcI)

	copySrcVertRange := func(fromI, toI int) FaceT {
		from, to := dstIToSrcI[fromI], dstIToSrcI[toI]
		result := make(FaceT, 0, int(math.Abs(float64(from-to)))+1)
		for i := range srcFace {
			nextI := (i + from) % len(srcFace)
			result = append(result, srcFace[nextI])
			if nextI == to {
				break
			}
		}
		return result
	}

	for i, dstVertIdx := range dstFace {
		nextI := (i + 1) % len(dstFace)
		nextDstVertIdx := dstFace[nextI]
		newFace := append(FaceT{dstVertIdx, nextDstVertIdx}, copySrcVertRange(nextI, i)...)
		if i == 0 {
			fi.dst.faces[dstFaceIdx] = newFace // replace old one
			// log.Printf("replacing original dstFaceIdx with:\n%v", fi.m.dumpFace(dstFaceIdx, fi.dst.faces[dstFaceIdx]))
			continue
		}
		fi.dst.faces = append(fi.dst.faces, newFace)
		// log.Printf("adding new dst face:\n%v", fi.m.dumpFace(faceIndexT(len(fi.dst.faces)), newFace))
	}
}

func (fi *faceInfoT) findClosestFaceByCentroid(srcFace FaceT, dstFaceIndices []faceIndexT) faceIndexT {
	srcCenter := fi.m.faceCenter(srcFace)
	var bestDstFaceIdx faceIndexT
	var bestDist float64
	for i, dstFaceIdx := range dstFaceIndices {
		dstFace := fi.dst.faces[dstFaceIdx]
		dstCenter := fi.m.faceCenter(dstFace)
		dist := dstCenter.Sub(srcCenter).Length()
		if i == 0 || dist < bestDist {
			bestDstFaceIdx = dstFaceIdx
			bestDist = dist
		}
	}
	return bestDstFaceIdx
}
