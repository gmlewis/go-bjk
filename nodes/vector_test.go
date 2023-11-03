package nodes

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetRotXYZ(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		v      Vec3
		wantRX float64
		wantRY float64
		wantRZ float64
	}{
		{
			name: "empty vec3 - origin",
		},
		{
			name: "x axis",
			v:    Vec3{1, 0, 0},
		},
		{
			name:   "neg x axis",
			v:      Vec3{-1, 0, 0},
			wantRY: math.Pi,
			wantRZ: math.Pi,
		},
		{
			name:   "y axis",
			v:      Vec3{0, 1, 0},
			wantRZ: math.Pi / 2,
		},
		{
			name:   "neg y axis",
			v:      Vec3{0, -1, 0},
			wantRX: math.Pi,
			wantRZ: -math.Pi / 2,
		},
		{
			name:   "z axis",
			v:      Vec3{0, 0, 1},
			wantRX: math.Pi / 2,
			wantRY: math.Pi / 2,
		},
		{
			name:   "neg z axis",
			v:      Vec3{0, 0, -1},
			wantRX: -math.Pi / 2,
			wantRY: -math.Pi / 2,
		},
	}

	eq := func(v1, v2 float64) bool { return cmp.Equal(v1, v2, cmpopts.EquateApprox(0.00001, 0)) }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx, ry, rz := tt.v.GetRotXYZ()
			if !eq(tt.wantRX, rx) || !eq(tt.wantRY, ry) || !eq(tt.wantRZ, rz) {
				t.Errorf("GetRotXYZ = %v, want %v", Vec3{rx, ry, rz}, Vec3{tt.wantRX, tt.wantRY, tt.wantRZ})
			}
		})
	}
}
