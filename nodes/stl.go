// -*- compile-command: "go run ../examples/bifilar-electromagnet/main.go -o '' -stl ../out.stl"; -*-

package nodes

import (
	"errors"
	"fmt"

	"github.com/gmlewis/go-bjk/ast"
	"github.com/gmlewis/irmf-slicer/v3/stl"
)

// ToSTL "renders" a BJK design to a binary STL file.
func (c *Client) ToSTL(design *ast.BJK, filename string) error {
	if design == nil || design.Graph == nil {
		return errors.New("design missing graph")
	}

	mesh, err := c.Eval(design)
	if err != nil {
		return err
	}

	out, err := stl.New(filename)
	if err != nil {
		return err
	}

	for faceIndex := range mesh.Faces {
		if err := tesselateFace(out, mesh, faceIndex); err != nil {
			return err
		}
	}
	return out.Close()
}

type stlWriter interface {
	Write(t *stl.Tri) error
}

func (v Vec3) tof32arr() [3]float32 {
	return [3]float32{float32(v.X), float32(v.Y), float32(v.Z)}
}

func tesselateFace(out stlWriter, mesh *Mesh, faceIndex int) error {
	face := mesh.Faces[faceIndex]
	if len(face) < 3 {
		return fmt.Errorf("face <3 verts: %+v", face)
	}

	var faceNormal Vec3
	if faceIndex < len(mesh.FaceNormals) {
		faceNormal = mesh.FaceNormals[faceIndex]
	} else {
		faceNormal = mesh.CalcFaceNormal(faceIndex)
	}
	n := faceNormal.tof32arr()
	// v1 := mesh.Verts[face[0]].tof32arr()

	// for i := 2; i < len(face); i++ {
	// 	v2, v3 := mesh.Verts[face[i-1]].tof32arr(), mesh.Verts[face[i]].tof32arr()
	// 	t := &stl.Tri{N: n, V1: v1, V2: v2, V3: v3}
	// 	if err := out.Write(t); err != nil {
	// 		return err
	// 	}
	// }

	// Note that this function is sent convex and concave polygons.
	// The algorithms needed to handle these are a pain, so this function
	// just uses a simple heuristic to start at one vertex, and make a triangle
	// with the vertices on either side of it, then works its way on alternating
	// sides of the face to the other end.
	numVerts := len(face)
	lastCWIdx, lastCCWIdx := 0, 1
	for {
		v1 := mesh.Verts[face[lastCWIdx]].tof32arr()
		v2 := mesh.Verts[face[lastCCWIdx]].tof32arr()

		// advance CW
		lastCWIdx = (lastCWIdx - 1 + numVerts) % numVerts
		if lastCWIdx == lastCCWIdx {
			break
		}
		v3 := mesh.Verts[face[lastCWIdx]].tof32arr()
		t := &stl.Tri{
			N:  n,
			V1: v1,
			V2: v2,
			V3: v3,
		}
		if err := out.Write(t); err != nil {
			return err
		}

		// advance CCW
		lastCCWIdx = (lastCCWIdx + 1) % numVerts
		if lastCWIdx == lastCCWIdx {
			break
		}
		v4 := mesh.Verts[face[lastCCWIdx]].tof32arr()
		t = &stl.Tri{
			N:  n,
			V1: v2,
			V2: v4,
			V3: v3,
		}
		if err := out.Write(t); err != nil {
			return err
		}
	}

	return nil
}
