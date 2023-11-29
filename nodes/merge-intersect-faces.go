// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"log"
	"sort"

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

	edgePairDist := map[edgeT]float64{}
	srcFace, dstFace := fi.src.faces[srcFaceIdx], fi.dst.faces[dstFaceIdx]
	edgeKeys := make([]edgeT, 0, len(srcFace)*len(dstFace))
	for _, srcVertIdx := range srcFace {
		for _, dstVertIdx := range dstFace {
			edge := makeEdge(srcVertIdx, dstVertIdx)
			edgeKeys = append(edgeKeys, edge)
			dist := fi.m.Verts[srcVertIdx].Sub(fi.m.Verts[dstVertIdx]).Length()
			edgePairDist[edge] = dist
		}
	}
	sort.Slice(edgeKeys, func(i, j int) bool { return edgePairDist[edgeKeys[i]] < edgePairDist[edgeKeys[j]] })

	for i, edge := range edgeKeys {
		log.Printf("edge #%v of %v: %v: dist=%0.2f", i+1, len(edgeKeys), edge, edgePairDist[edge])
	}
}
