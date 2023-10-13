package nodes

import (
	_ "embed"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	// TODO: make this on a machine other than my laptop?!?
	repoDir = "/Users/glenn/src/github.com/gmlewis/blackjack"
)

//go:embed testdata/bifilar-electromagnet.bjk
var bifilarElectromagnet string

func TestBuild(t *testing.T) {
	c, err := New(repoDir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	design, err := c.NewBuilder().
		AddNode("MakeScalar.vert-turns", "x=2").
		AddNode("Point.helix-bbox", "point=vector(3,2,3)").
		AddNode("VectorMath.vert-gap", "vec_b=vector(0,0.5,0)").
		Connect("Point.helix-bbox.point", "VectorMath.vert-gap.point").
		AddNode("Helix.wire-1", "start_angle=180", "segments=36").
		Connect("VectorMath.vert-gap.out", "Helix.wire-1.size").
		AddNode("Helix.wire-2", "start_angle=0", "segments=36").
		AddNode("MakeQuad.wire-outline", "size=vector(1,1,1)").
		AddNode("ExtrudeAlongCurve.wire-1").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-1.cross_section").
		Connect("Helix.wire-1.out_mesh", "ExtrudeAlongCurve.wire-1.backbone").
		AddNode("ExtrudeAlongCurve.wire-2").
		Connect("MakeQuad.wire-outline.out_mesh", "ExtrudeAlongCurve.wire-2.cross_section").
		Connect("Helix.wire-2.out_mesh", "ExtrudeAlongCurve.wire-2.backbone").
		AddNode("MergeMeshes.wire-1-2").
		Connect("ExtrudeAlongCurve.wire-1.out_mesh", "MergeMeshes.wire-1-2.mesh_a").
		Connect("ExtrudeAlongCurve.wire-2.out_mesh", "MergeMeshes.wire-1-2.mesh_b").
		Build()
	if err != nil {
		t.Fatal(err)
	}

	got, want := design.String(), bifilarElectromagnet
	if diff := cmp.Diff(want, got); diff != "" {
		t.Log(got)
		t.Errorf("design mismatch (-want +got):\n%v", diff)
	}
}
