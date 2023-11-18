package nodes

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/maps"
)

func TestBadEdgesToConnectedEdgeLoops(t *testing.T) {
	tests := []struct {
		name     string
		badEdges []edgeT
		want     []faceKeyT
	}{
		{
			name: "empty",
			want: []faceKeyT{},
		},
		{
			name: "one edge loop",
			badEdges: []edgeT{
				makeEdge(0, 1),
				makeEdge(2, 3),
				makeEdge(3, 0),
				makeEdge(1, 2),
			},
			want: []faceKeyT{"[0 1 2 3]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &infoSetT{
				badEdges: edgeToFacesMapT{},
			}
			for _, edge := range tt.badEdges {
				is.badEdges[edge] = nil // value doesn't matter, key does.
			}

			gotMap := is.badEdgesToConnectedEdgeLoops()
			got := maps.Keys(gotMap)
			sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("badEdgesToConnectedEdgeLoops mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
