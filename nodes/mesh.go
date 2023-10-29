package nodes

// Mesh represents a mesh of points, edges, and faces.
type Mesh struct {
	Verts []Vec3
	Faces [][]int
}

// NewMeshFromPolygons creates a new mesh from points.
func NewMeshFromPolygons(verts []Vec3, faces [][]int) *Mesh {
	return &Mesh{Verts: verts, Faces: faces}
}
