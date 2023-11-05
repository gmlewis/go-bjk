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

func tesselateFace(out *stl.Client, mesh *Mesh, faceIndex int) error {
	face := mesh.Faces[faceIndex]
	if len(face) < 3 {
		return fmt.Errorf("face <3 verts: %+v", face)
	}
	faceNormal := mesh.CalcFaceNormal(faceIndex)
	n := [3]float32{float32(faceNormal.X), float32(faceNormal.Y), float32(faceNormal.Z)}
	pt1 := mesh.Verts[face[0]]
	v1 := [3]float32{float32(pt1.X), float32(pt1.Y), float32(pt1.Z)}
	for i := 2; i < len(face); i++ {
		pt2, pt3 := mesh.Verts[face[i-1]], mesh.Verts[face[i]]
		v2 := [3]float32{float32(pt2.X), float32(pt2.Y), float32(pt2.Z)}
		v3 := [3]float32{float32(pt3.X), float32(pt3.Y), float32(pt3.Z)}
		t := &stl.Tri{N: n, V1: v1, V2: v2, V3: v3}
		if err := out.Write(t); err != nil {
			return err
		}
	}

	return nil
}
