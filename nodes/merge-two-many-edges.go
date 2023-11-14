package nodes

import (
	"fmt"
	"log"

	"golang.org/x/exp/maps"
)

func (fi *faceInfoT) merge2manisManyEdges(sharedVerts sharedVertsMapT, sharedEdges sharedEdgesMapT) {
	// first, generate a map of edge heights for all src edges and all dst edges.
	srcEdgeHeights := make(map[edgeT]float64, len(sharedEdges))
	dstEdgeHeights := make(map[edgeT]float64, len(sharedEdges))
	uniqueHeights := map[string]float64{}
	for edge, v := range sharedEdges {
		srcHeight := fi.src.edgeHeight(edge)
		dstHeight := fi.dst.edgeHeight(edge)
		log.Printf("merge2manisManyEdges: shared edge %v: src faces: %+v (height %0.2f), dst faces: %+v (height %0.2f)", edge, v[0], srcHeight, v[1], dstHeight)
		srcEdgeHeights[edge] = srcHeight
		dstEdgeHeights[edge] = dstHeight

		uniqueHeights[fmt.Sprintf("%0.5f", srcHeight)] = srcHeight
		uniqueHeights[fmt.Sprintf("%0.5f", dstHeight)] = dstHeight
	}
	log.Printf("merge2manisManyEdges: %v unique heights: %+v", len(uniqueHeights), maps.Keys(uniqueHeights))

	if len(uniqueHeights) == 1 {
		heights := maps.Values(uniqueHeights)
		fi.merge2manisManyEdgesOneHeight(sharedVerts, sharedEdges, heights[0])
		return
	}

	log.Fatalf("merge2manisManyEdges: not implemented yet")
}

func (fi *faceInfoT) merge2manisManyEdgesOneHeight(sharedVerts sharedVertsMapT, sharedEdges sharedEdgesMapT, height float64) {
	for edge, v := range sharedEdges {
		srcFaces := v[0]
		dstFaces := v[1]
		log.Printf("merge2manisManyEdgesOneHeight: edge %v, src face[%v] normal: %v", edge, srcFaces[0], fi.src.faceNormals[srcFaces[0]])
		log.Printf("merge2manisManyEdgesOneHeight: edge %v, src face[%v] normal: %v", edge, srcFaces[1], fi.src.faceNormals[srcFaces[1]])
		log.Printf("merge2manisManyEdgesOneHeight: edge %v, dst face[%v] normal: %v", edge, dstFaces[0], fi.dst.faceNormals[dstFaces[0]])
		log.Printf("merge2manisManyEdgesOneHeight: edge %v, dst face[%v] normal: %v", edge, dstFaces[1], fi.dst.faceNormals[dstFaces[1]])
	}

	log.Fatalf("merge2manisManyEdgesOneHeight: not implemented yet")
}
