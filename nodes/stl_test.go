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
	if err := tesselateFace(out, mesh, 0); err != nil {
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
		V1: [3]float32{11.5, -0.5, 0},
		V2: [3]float32{10.199999809265137, -0.5, 5.309999942779541},
		V3: [3]float32{12.5, -0.5, 0},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{10.199999809265137, -0.5, 5.309999942779541},
		V2: [3]float32{6.590000152587891, -0.5, 9.420000076293945},
		V3: [3]float32{12.5, -0.5, 0},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{12.5, -0.5, 0},
		V2: [3]float32{6.590000152587891, -0.5, 9.420000076293945},
		V3: [3]float32{11.069999694824219, -0.5, 5.809999942779541},
	},
	{
		N:  [3]float32{0, 1, 0},
		V1: [3]float32{6.590000152587891, -0.5, 9.420000076293945},
		V2: [3]float32{7.099999904632568, -0.5, 10.289999961853027},
		V3: [3]float32{11.069999694824219, -0.5, 5.809999942779541},
	},
}
