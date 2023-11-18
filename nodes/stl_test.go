package nodes

import (
	"testing"

	"github.com/gmlewis/irmf-slicer/v3/stl"
	"github.com/google/go-cmp/cmp"
)

func TestTesselateFace(t *testing.T) {
	mesh := NewPolygonFromPoints(
		[]Vec3{
			{11.50, -0.50, 0.00},
			{10.20, -0.50, 5.31},
			{6.59, -0.50, 9.42},
			{7.10, -0.50, 10.29},
			{11.07, -0.50, 5.81},
			{12.50, -0.50, 0.00},
		})

	out := &fakeSTLWriter{}
	if err := tesselateFace(out, mesh, 0, false); err != nil {
		t.Fatal(err)
	}

	got := out.tris
	if diff := cmp.Diff(wantTris, got); diff != "" {
		t.Errorf("tesselateFace mismatch (-want +got):\n%v", diff)
	}
}

type fakeSTLWriter struct {
	tris []*stl.Tri
}

func (f *fakeSTLWriter) Write(t *stl.Tri) error {
	f.tris = append(f.tris, t)
	return nil
}

var wantTris = []*stl.Tri{
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{12.5, -0.5, 0},
		V2: [3]float32{11.5, -0.5, 0},
		V3: [3]float32{11.07, -0.50, 5.81},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{11.5, -0.5, 0},
		V2: [3]float32{10.20, -0.50, 5.31},
		V3: [3]float32{11.07, -0.50, 5.81},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{11.07, -0.50, 5.81},
		V2: [3]float32{10.20, -0.50, 5.31},
		V3: [3]float32{7.10, -0.50, 10.29},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{10.20, -0.50, 5.31},
		V2: [3]float32{6.59, -0.50, 9.42},
		V3: [3]float32{7.10, -0.50, 10.29},
	},
}
