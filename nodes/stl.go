// -*- compile-command: "go run ../examples/bifilar-electromagnet/main.go -o '' -stl ../out.stl"; -*-

package nodes

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/fogleman/fauxgl"
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

	return mesh.WriteSTL(filename)
}

// WriteSTL writes the mesh to a new STL file.
func (m *Mesh) WriteSTL(filename string) error {
	out, err := stl.New(filename)
	if err != nil {
		return err
	}

	for faceIndex := range m.Faces {
		if err := tesselateFace(out, m, faceIndex); err != nil {
			return err
		}
	}
	return out.Close()
}

// STLToMesh reads an STL file and returns a new (triangulated) Mesh.
func STLToMesh(filename string) (*Mesh, error) {
	mesh, err := fauxgl.LoadSTL(filename)
	if err != nil {
		return nil, err
	}

	m := NewMesh()
	for _, tri := range mesh.Triangles {
		v1, v2, v3 := tri.V1.Position, tri.V2.Position, tri.V3.Position
		verts := []Vec3{
			Vec3{X: v1.X, Y: v1.Y, Z: v1.Z},
			Vec3{X: v2.X, Y: v2.Y, Z: v2.Z},
			Vec3{X: v3.X, Y: v3.Y, Z: v3.Z},
		}
		m.AddFace(verts)
	}
	return m, nil
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

	faceNormal := mesh.CalcFaceNormal(mesh.Faces[faceIndex])
	n := faceNormal.tof32arr()

	// Note that this function is sent convex and concave polygons.
	// This simple algorithm does not always choose the face with the smallest
	// area. Therefore, run it twice with two different starting points, and
	// chose the face with the smallest final area.
	numVerts := len(face)
	faceTris, faceArea := simpleTesselator(mesh.Verts, n, face, 0, 1)
	// log.Printf("face1Area=%0.5f", faceArea)
	if faceTris2, face2Area := simpleTesselator(mesh.Verts, n, face, 1, 2); face2Area < faceArea {
		// log.Printf("face2Area=%0.5f: face2 WINS!", face2Area)
		faceTris, faceArea = faceTris2, face2Area
		// 	} else {
		// 		log.Printf("face2Area=%0.5f: face1 wins", face2Area)
	}
	if numVerts > 3 {
		if faceTris3, face3Area := simpleTesselator(mesh.Verts, n, face, numVerts-1, 0); face3Area < faceArea {
			//			log.Printf("face3Area=%0.5f: face3 WINS!", face3Area)
			faceTris, faceArea = faceTris3, face3Area
			//		} else {
			//			log.Printf("face3Area=%0.5f: face1 wins", face3Area)
		}
	}

	for _, t := range faceTris {
		if err := out.Write(t); err != nil {
			return err
		}
	}

	return nil
}

// The algorithms needed to handle these are a pain, so this function
// just uses a simple heuristic to start at one vertex, and make a triangle
// with the vertices on either side of it, then works its way on alternating
// sides of the face to the other end.
func simpleTesselator(verts []Vec3, n [3]float32, face FaceT, lastCWIdx, lastCCWIdx int) ([]*stl.Tri, float64) {
	numVerts := len(face)
	faceTris := make([]*stl.Tri, 0, numVerts-2)
	var totalArea float64

	for {
		v1 := verts[face[lastCWIdx]]
		v1f32 := v1.tof32arr()
		v2 := verts[face[lastCCWIdx]]
		v2f32 := v2.tof32arr()

		// advance CW
		lastCWIdx = (lastCWIdx - 1 + numVerts) % numVerts
		if lastCWIdx == lastCCWIdx {
			break
		}
		v3 := verts[face[lastCWIdx]]
		v3f32 := v3.tof32arr()
		t := &stl.Tri{
			N:  n,
			V1: v1f32,
			V2: v2f32,
			V3: v3f32,
		}
		faceTris = append(faceTris, t)
		totalArea += triangleArea(v1, v2, v3)

		// advance CCW
		lastCCWIdx = (lastCCWIdx + 1) % numVerts
		if lastCWIdx == lastCCWIdx {
			break
		}
		v4 := verts[face[lastCCWIdx]]
		v4f32 := v4.tof32arr()
		t = &stl.Tri{
			N:  n,
			V1: v2f32,
			V2: v4f32,
			V3: v3f32,
		}
		faceTris = append(faceTris, t)
		totalArea += triangleArea(v2, v4, v3)
	}

	return faceTris, totalArea
}

// https://math.stackexchange.com/questions/128991/how-to-calculate-the-area-of-a-3d-triangle
func triangleArea(va, vb, vc Vec3) float64 {
	vab := va.Sub(vb)
	abLength := vab.Length()
	if abLength == 0 {
		log.Printf("triangleArea(%v,%v,%v) area is zero!", va, vb, vc)
		return 0
	}
	vac := va.Sub(vc)
	acLength := vac.Length()
	if acLength == 0 {
		log.Printf("triangleArea(%v,%v,%v) area is zero!", va, vb, vc)
		return 0
	}
	cosTheta := Vec3Dot(vab, vac) / (abLength * acLength)
	theta := math.Acos(cosTheta)

	return 0.5 * abLength * acLength * math.Sin(theta)
}
