package nodes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCalcFaceNormal(t *testing.T) {
	mesh := NewPolygonFromPoints(
		[]Vec3{
			{11.50, -0.50, 0.00},
			{10.20, -0.50, 5.31},
			{6.59, -0.50, 9.42},
			{7.10, -0.50, 10.29},
			{11.07, -0.50, 5.81},
			{12.50, -0.50, 0.00},
		})

	got := mesh.CalcFaceNormal(mesh.Faces[0])
	want := Vec3{0, 1, 0}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CalcFaceNormal mismatch (-want +got):\n%v", diff)
	}
}
