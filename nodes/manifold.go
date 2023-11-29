// -*- compile-command: "go test -v ./..."; -*-

package nodes

import (
	"fmt"
	"log"
	"math"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// edgeT represents an edge and is a sorted array of two vertex indices.
type edgeT [2]VertIndexT

// edgeToFacesMapT represents a mapping from an edge to one or more face indices.
type edgeToFacesMapT map[edgeT][]faceIndexT

// vertToFacesMapT respresents a mapping from a vertex index to face indices.
type vertToFacesMapT map[VertIndexT][]faceIndexT

// face2EdgesMapT represents a mapping from a face index to edges.
type face2EdgesMapT map[faceIndexT][]edgeT

// faceStrToFaceIdxMapT maps a face "signature" (e.g. "0 1 2 3") to a face index.
type faceStrToFaceIdxMapT map[faceKeyT]faceIndexT

// sharedVertsMapT represents a collection of shared vertices and maps them back to src ([0]) and dst([1]) face indices.
type sharedVertsMapT map[VertIndexT][2][]faceIndexT

// sharedEdgesMapT represents a collection of shared edges and maps them back to src ([0]) and dst([1]) face indices.
type sharedEdgesMapT map[edgeT][2][]faceIndexT

// sharedFacesMapT represents a collection of shared faces (keyed by face "signature") and maps them back to
// src ([0]) and dst([1]) face index.
type sharedFacesMapT map[faceKeyT][2]faceIndexT

type faceInfoT struct {
	m   *Mesh
	src *infoSetT
	dst *infoSetT
}

type infoSetT struct {
	faceInfo         *faceInfoT
	faces            []FaceT
	faceNormals      []Vec3
	vertToFaces      vertToFacesMapT
	edgeToFaces      edgeToFacesMapT
	faceStrToFaceIdx faceStrToFaceIdxMapT
	badEdges         edgeToFacesMapT
	badFaces         face2EdgesMapT

	facesTargetedForDeletion map[faceIndexT]bool
}

func makeEdge(v1, v2 VertIndexT) edgeT {
	if v1 == v2 {
		log.Fatalf("programming error: makeEdge(%v,%v)", v1, v2)
	}
	if v1 < v2 {
		return edgeT{v1, v2}
	}
	return edgeT{v2, v1} // swap
}

func makeFaceKeyFromEdges(edges []edgeT) faceKeyT {
	verts := map[VertIndexT]struct{}{}
	for _, edge := range edges {
		verts[edge[0]] = struct{}{}
		verts[edge[1]] = struct{}{}
	}

	face := FaceT(maps.Keys(verts))
	return face.toKey()
}

// genFaceInfo calculates the face normals for every src and dst face
// and generates a map of good and bad edges (mapped to their respective faces).
func (m *Mesh) genFaceInfo(dstFaces, srcFaces []FaceT) *faceInfoT {
	fi := &faceInfoT{m: m}
	fi.src = fi.genFaceInfoForSet(srcFaces)
	fi.dst = fi.genFaceInfoForSet(dstFaces)
	return fi
}

func (fi *faceInfoT) genFaceInfoForSet(faces []FaceT) *infoSetT {
	infoSet := &infoSetT{
		faceInfo:         fi,
		faces:            faces,
		faceNormals:      make([]Vec3, 0, len(faces)),
		vertToFaces:      vertToFacesMapT{}, // key=vertIdx, value=[]faceIdx
		edgeToFaces:      edgeToFacesMapT{},
		faceStrToFaceIdx: faceStrToFaceIdxMapT{},
		badEdges:         edgeToFacesMapT{},
		badFaces:         face2EdgesMapT{},

		facesTargetedForDeletion: map[faceIndexT]bool{},
	}

	for i, face := range faces {
		faceIdx := faceIndexT(i)
		infoSet.faceNormals = append(infoSet.faceNormals, fi.m.CalcFaceNormal(face))
		infoSet.faceStrToFaceIdx[face.toKey()] = faceIdx
		for j, vertIdx := range face {
			infoSet.vertToFaces[vertIdx] = append(infoSet.vertToFaces[vertIdx], faceIdx)
			nextVertIdx := face[(j+1)%len(face)]
			edge := makeEdge(vertIdx, nextVertIdx)
			infoSet.edgeToFaces[edge] = append(infoSet.edgeToFaces[edge], faceIdx)
		}
	}

	// Now find the bad edges and move them to the badEdges map.
	for edge, faceIdxes := range infoSet.edgeToFaces {
		if len(faceIdxes) != 2 {
			infoSet.badEdges[edge] = faceIdxes
			for _, faceIdx := range faceIdxes {
				infoSet.badFaces[faceIdx] = append(infoSet.badFaces[faceIdx], edge)
			}
		}
	}
	for edge := range infoSet.badEdges {
		delete(infoSet.edgeToFaces, edge)
	}

	return infoSet
}

func (fi *faceInfoT) findSharedVEFs() (sharedVertsMapT, sharedEdgesMapT, sharedFacesMapT) {
	// premature optimization:
	// if len(fi.dstFaces) < len(fi.srcFaces) {
	// 	fi.swapSrcAndDst()
	// }

	sharedVerts := sharedVertsMapT{}
	for vertIdx, dstFaces := range fi.dst.vertToFaces {
		if srcFaces, ok := fi.src.vertToFaces[vertIdx]; ok {
			sharedVerts[vertIdx] = [2][]faceIndexT{srcFaces, dstFaces}
		}
	}

	sharedEdges := sharedEdgesMapT{}
	for edge, dstFaces := range fi.dst.edgeToFaces {
		if srcFaces, ok := fi.src.edgeToFaces[edge]; ok {
			sharedEdges[edge] = [2][]faceIndexT{srcFaces, dstFaces}
		}
	}

	sharedFaces := sharedFacesMapT{}
	for faceStr, dstFaceIdx := range fi.dst.faceStrToFaceIdx {
		if srcFaceIdx, ok := fi.src.faceStrToFaceIdx[faceStr]; ok {
			sharedFaces[faceStr] = [2]faceIndexT{srcFaceIdx, dstFaceIdx}
		}
	}

	return sharedVerts, sharedEdges, sharedFaces
}

type edgeVectorT struct {
	edge        edgeT
	fromVertIdx VertIndexT
	toVertIdx   VertIndexT
	toSubFrom   Vec3
	length      float64
}

func (ev edgeVectorT) String() string {
	return fmt.Sprintf("{from:%v to:%v %v length:%0.2f}", ev.fromVertIdx, ev.toVertIdx, ev.toSubFrom, ev.length)
}

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on `edge`)
// and therefore is completely independent of the winding order of the face!
// Both edges are found in the `badEdges` map.
// In addition to the edge vector, it also returns the VertIndexT of the other vertex.
func (is *infoSetT) connectedBadEdgeVectorFromVert(vertIdx VertIndexT, edge edgeT) edgeVectorT {
	notVertIdx := edge[0]
	if notVertIdx == vertIdx {
		notVertIdx = edge[1]
	}

	for otherEdge := range is.badEdges {
		var nextIdx VertIndexT
		switch {
		case otherEdge[0] == vertIdx && otherEdge[1] != notVertIdx:
			nextIdx = otherEdge[1]
		case otherEdge[1] == vertIdx && otherEdge[0] != notVertIdx:
			nextIdx = otherEdge[0]
		default:
			continue
		}

		return is.faceInfo.m.makeEdgeVector(vertIdx, nextIdx)
	}

	log.Fatalf("connectedBadEdgeVectorFromVert: programming error for edge %v", edge)
	return edgeVectorT{}
}

func (m *Mesh) makeEdgeVector(fromIdx, toIdx VertIndexT) edgeVectorT {
	toSubFrom := m.Verts[toIdx].Sub(m.Verts[fromIdx])
	return edgeVectorT{
		edge:        makeEdge(fromIdx, toIdx),
		fromVertIdx: fromIdx,
		toVertIdx:   toIdx,
		toSubFrom:   toSubFrom,
		length:      toSubFrom.Length(),
	}
}

// otherVertexFrom returns the other vertex connected to this edge starting at vertIdx on the given faceIdx.
func (is *infoSetT) otherVertexFrom(edge edgeT, vertIdx VertIndexT, faceIdx faceIndexT) VertIndexT {
	face := is.faces[faceIdx]
	for i, vIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if makeEdge(vIdx, nextIdx) == edge {
			continue
		}
		if vIdx == vertIdx {
			return nextIdx
		}
		if nextIdx == vertIdx {
			return vIdx
		}
	}
	log.Fatalf("otherVertexFrom: programming error for edge %v, vertIdx=%v, faceIdx=%v", edge, vertIdx, faceIdx)
	return 0
}

// makeEdgeVectors returns two edgeVectorTs for the given faceIdx, one for the first vertex, and one for the second.
func (is *infoSetT) makeEdgeVectors(edge edgeT, faceIdx faceIndexT) [2]edgeVectorT {
	m := is.faceInfo.m
	return [2]edgeVectorT{
		m.makeEdgeVector(edge[0], is.otherVertexFrom(edge, edge[0], faceIdx)),
		m.makeEdgeVector(edge[1], is.otherVertexFrom(edge, edge[1], faceIdx)),
	}
}

// makeEdgeVectorsFromVert returns two edgeVectorTs for the given faceIdx, both from the vertex.
func (is *infoSetT) makeEdgeVectorsFromVert(vertIdx VertIndexT, faceIdx faceIndexT) [2]edgeVectorT {
	m := is.faceInfo.m
	face := is.faces[faceIdx]
	for i, vIdx := range face {
		lastIdx := face[(i-1+len(face))%len(face)]
		nextIdx := face[(i+1)%len(face)]
		if vIdx == vertIdx {
			return [2]edgeVectorT{
				m.makeEdgeVector(vIdx, lastIdx),
				m.makeEdgeVector(vIdx, nextIdx),
			}
		}
	}
	log.Fatalf("makeEdgeVectorsFromVert: programming error")
	return [2]edgeVectorT{}
}

// Note that this vector is pointing FROM vertIdx TOWARD the other connected vertex (not on `edge`)
// and therefore is completely independent of the winding order of the face!
// In addition to the edge vector, it also returns the VertIndexT of the other vertex.
func (is *infoSetT) connectedEdgeVectorFromVertOnFace(vertIdx VertIndexT, edge edgeT, faceIdx faceIndexT) edgeVectorT {
	notVertIdx := edge[0]
	if notVertIdx == vertIdx {
		notVertIdx = edge[1]
	}

	m := is.faceInfo.m
	face := is.faces[faceIdx]
	for i, pIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		if pIdx == vertIdx && nextIdx != notVertIdx {
			// log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, nextIdx=%v, returning (%v).Sub(%v)",
			//   vertIdx, edge, faceIdx, i, pIdx, nextIdx, m.Verts[nextIdx], m.Verts[vertIdx])
			toSubFrom := m.Verts[nextIdx].Sub(m.Verts[vertIdx])
			return edgeVectorT{
				edge:        makeEdge(vertIdx, nextIdx),
				fromVertIdx: vertIdx,
				toVertIdx:   nextIdx,
				toSubFrom:   toSubFrom,
				length:      toSubFrom.Length(),
			}
		}
		if pIdx == vertIdx {
			lastVertIdx := face[(i-1+len(face))%len(face)]
			// log.Printf("connectedEdgeVectorFromVertOnFace(vertIdx=%v, edge=%v, faceIdx=%v): i=%v, pIdx=%v, lastVertIdx=%v, returning (%v).Sub(%v)",
			//   vertIdx, edge, faceIdx, i, pIdx, lastVertIdx, m.Verts[lastVertIdx], m.Verts[vertIdx])
			toSubFrom := m.Verts[lastVertIdx].Sub(m.Verts[vertIdx])
			return edgeVectorT{
				edge:        makeEdge(vertIdx, lastVertIdx),
				fromVertIdx: vertIdx,
				toVertIdx:   lastVertIdx,
				toSubFrom:   toSubFrom,
				length:      toSubFrom.Length(),
			}
		}
	}

	log.Fatalf("connectedEdgeVectorFromVertOnFace: programming error for face %+v", face)
	return edgeVectorT{}
}

// moveVerts creates new (or reuses old) vertices and returns the mapping from the
// old face's vertIndexes to the new vertices, without modifying the face.
func (is *infoSetT) moveVerts(face FaceT, move Vec3) vToVMap {
	m := is.faceInfo.m

	vertsOldToNew := make(vToVMap, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx].Add(move)
		newVertIdx := m.AddVert(v)
		vertsOldToNew[vertIdx] = newVertIdx
	}

	return vertsOldToNew
}

type vToVMap map[VertIndexT]VertIndexT

// getFaceSideEdgeVectors returns a slice of edge vectors that are connected to (but not on) this face.
func (is *infoSetT) getFaceSideEdgeVectors(baseFaceIdx faceIndexT) []edgeVectorT {
	face := is.faces[baseFaceIdx]
	result := make([]edgeVectorT, 0, len(face))
	for i, vertIdx := range face {
		nextIdx := face[(i+1)%len(face)]
		edge := makeEdge(vertIdx, nextIdx)
		facesFromEdge := is.edgeToFaces[edge]
		for _, otherFaceIdx := range facesFromEdge {
			if otherFaceIdx == baseFaceIdx {
				continue
			}
			// take one edge from each connected face
			ev := is.connectedEdgeVectorFromVertOnFace(vertIdx, edge, otherFaceIdx)
			result = append(result, ev)
			break
		}
	}

	return result
}

func reverseMapFaceIndicesToEdges(sharedEdges sharedEdgesMapT) (srcFaceIndicesToEdges, dstFaceIndicesToEdges face2EdgesMapT) {
	srcFaceIndicesToEdges, dstFaceIndicesToEdges = face2EdgesMapT{}, face2EdgesMapT{}
	for edge, v := range sharedEdges {
		for _, faceIdx := range v[0] {
			srcFaceIndicesToEdges[faceIdx] = append(srcFaceIndicesToEdges[faceIdx], edge)
		}
		for _, faceIdx := range v[1] {
			dstFaceIndicesToEdges[faceIdx] = append(dstFaceIndicesToEdges[faceIdx], edge)
		}
	}
	return srcFaceIndicesToEdges, dstFaceIndicesToEdges
}

// faceIndicesByEdgeCount returns a map of edge count to slice of faceIndices.
// So a face that has 6 shared edges would appear in the slice in result[6].
func faceIndicesByEdgeCount(inMap face2EdgesMapT) map[int][]faceIndexT {
	result := map[int][]faceIndexT{}
	for faceIdx, edges := range inMap {
		result[len(edges)] = append(result[len(edges)], faceIdx)
	}
	return result
}

// deleteFace deletes the face at the provided index, thereby shifting the other
// face indices around it! Always delete from last to first when deleting multiple faces.
// Do not call this directly. Let deleteFacesLastToFirst actually delete the faces.
func (is *infoSetT) deleteFace(deleteFaceIdx faceIndexT) {
	// log.Printf("\n\nDELETING FACE!!! %v", is.faceInfo.m.dumpFace(deleteFaceIdx, is.faces[deleteFaceIdx]))
	is.faces = slices.Delete(is.faces, int(deleteFaceIdx), int(deleteFaceIdx+1)) // invalidates other faceInfoT maps - last step.
}

// deleteFacesLastToFirst deletes faces by sorting their indices, then deleting them highest to lowest.
func (is *infoSetT) deleteFacesLastToFirst(facesToDeleteMap map[faceIndexT]bool) {
	facesToDelete := maps.Keys(facesToDeleteMap)
	sort.Slice(facesToDelete, func(i, j int) bool { return facesToDelete[i] > facesToDelete[j] })
	for _, faceIdx := range facesToDelete {
		// if is == is.faceInfo.src {
		// 	log.Printf("*** Deleting src face: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
		// } else {
		// 	log.Printf("*** Deleting dst face: %v", is.faceInfo.m.dumpFace(faceIdx, is.faces[faceIdx]))
		// }
		is.deleteFace(faceIdx)
	}
}

func (is *infoSetT) otherFaceOnEdge(edge edgeT, otherFaceIdx faceIndexT) (faceIndexT, edgeT, edgeT) {
	for _, faceIdx := range is.edgeToFaces[edge] {
		if faceIdx == otherFaceIdx {
			continue
		}
		v0 := is.otherVertexFrom(edge, edge[0], faceIdx)
		v1 := is.otherVertexFrom(edge, edge[1], faceIdx)
		return faceIdx, makeEdge(edge[0], v0), makeEdge(edge[1], v1)
	}
	log.Fatalf("otherFaceOnEdge: programming error")
	return 0, edgeT{}, edgeT{}
}

func (m *Mesh) faceArea(face FaceT) float64 {
	if len(face) == 4 {
		v1 := m.Verts[face[1]].Sub(m.Verts[face[0]]).Length()
		v2 := m.Verts[face[2]].Sub(m.Verts[face[1]]).Length()
		return v1 * v2
	}

	p1 := m.Verts[face[0]]
	p2 := m.Verts[face[1]]
	p3 := m.Verts[face[2]]
	a := math.Pow(((p2.Y-p1.Y)*(p3.Z-p1.Z)-(p3.Y-p1.Y)*(p2.Z-p1.Z)), 2) + math.Pow(((p3.X-p1.X)*(p2.Z-p1.Z)-(p2.X-p1.X)*(p3.Z-p1.Z)), 2) + math.Pow(((p2.X-p1.X)*(p3.Y-p1.Y)-(p3.X-p1.X)*(p2.Y-p1.Y)), 2)
	cosnx := ((p2.Y-p1.Y)*(p3.Z-p1.Z) - (p3.Y-p1.Y)*(p2.Z-p1.Z)) / math.Sqrt(a)
	cosny := ((p3.X-p1.X)*(p2.Z-p1.Z) - (p2.X-p1.X)*(p3.Z-p1.Z)) / math.Sqrt(a)
	cosnz := ((p2.X-p1.X)*(p3.Y-p1.Y) - (p3.X-p1.X)*(p2.Y-p1.Y)) / math.Sqrt(a)
	var s float64
	for i, vertIdx := range face {
		p1 = m.Verts[vertIdx]
		p2 = m.Verts[face[(i+1)%len(face)]]
		s += cosnz*((p1.X)*(p2.Y)-(p2.X)*(p1.Y)) + cosnx*((p1.Y)*(p2.Z)-(p2.Y)*(p1.Z)) + cosny*((p1.Z)*(p2.X)-(p2.Z)*(p1.X))
	}

	return math.Abs(0.5 * s)
}

func (fi *faceInfoT) closestVertOnFace(vertIdx VertIndexT, face FaceT) (int, VertIndexT) {
	var bestVertIdx VertIndexT
	var bestDist float64
	var bestI int
	refVert := fi.m.Verts[vertIdx]
	for i, vIdx := range face {
		dist := fi.m.Verts[vIdx].Sub(refVert).Length()
		if i == 0 || dist < bestDist {
			bestI = i
			bestDist = dist
			bestVertIdx = vIdx
		}
	}
	return bestI, bestVertIdx
}

// func (m *Mesh) dumpFaces(faces []FaceT) string {
// 	var lines []string
// 	for i, face := range faces {
// 		lines = append(lines, m.dumpFace(faceIndexT(i), face))
// 	}
// 	return strings.Join(lines, "\n")
// }

func (m *Mesh) dumpFacesByIndices(faceIndices []faceIndexT) string {
	var lines []string
	for _, faceIdx := range faceIndices {
		face := m.Faces[faceIdx]
		lines = append(lines, m.dumpFace(faceIdx, face))
	}
	return strings.Join(lines, "\n")
}

func (is *infoSetT) dumpFacesByIndices(faceIndices []faceIndexT) string {
	var lines []string
	for _, faceIdx := range faceIndices {
		face := is.faces[faceIdx]
		lines = append(lines, is.faceInfo.m.dumpFace(faceIdx, face))
	}
	return strings.Join(lines, "\n")
}

func (m *Mesh) dumpFace(faceIdx faceIndexT, face FaceT) string {
	verts := make([]string, 0, len(face))
	for _, vertIdx := range face {
		v := m.Verts[vertIdx]
		verts = append(verts, v.String())
	}
	return fmt.Sprintf("face[%v]={%+v}: {%v}", faceIdx, face, strings.Join(verts, " "))
}
