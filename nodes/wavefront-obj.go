package nodes

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ObjStrToMesh converts a simple Wavefront obj file
// (passed as a string) to a Mesh. Note that it only
// supports the bare minimum verts and faces.
//
// See: https://en.wikipedia.org/wiki/Wavefront_.obj_file
func ObjStrToMesh(objData string) (*Mesh, error) {
	var maxVertIdx int
	m := NewMesh()
	for i, line := range strings.Split(objData, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "", strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "l "):
		case strings.HasPrefix(line, "vn "):
		case strings.HasPrefix(line, "vp "):
		case strings.HasPrefix(line, "vt "):
		case strings.HasPrefix(line, "v "):
			parts := strings.Split(line, " ")
			if len(parts) < 4 {
				return nil, fmt.Errorf("unable to parse Wavefront obj file line #%v: %v", i+1, line)
			}
			x, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return nil, err
			}
			y, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return nil, err
			}
			z, err := strconv.ParseFloat(parts[3], 64)
			if err != nil {
				return nil, err
			}
			vertIdx := m.AddVert(Vec3{X: x, Y: y, Z: z})
			if int(vertIdx) > maxVertIdx {
				maxVertIdx = int(vertIdx)
			}
		case strings.HasPrefix(line, "f "):
			parts := strings.Split(line, " ")
			if len(parts) < 4 {
				return nil, fmt.Errorf("unable to parse Wavefront obj file line #%v: %v", i+1, line)
			}
			face := make(FaceT, 0, len(parts)-1)
			for _, s := range parts[1:] {
				if strings.Contains(s, "/") {
					p := strings.Split(s, "/")
					s = p[0]
				}
				v, err := strconv.Atoi(s)
				if err != nil {
					return nil, err
				}
				v-- // Wavefront obj is 1-indexed; mesh is 0-indexed.
				if v > maxVertIdx {
					return nil, fmt.Errorf("face '%v' has vertex index > max vertex index (%v)", line, maxVertIdx+1)
				}
				face = append(face, VertIndexT(v))
			}
			m.Faces = append(m.Faces, face)
		}
	}
	return m, nil
}

// WriteObj writes a mesh to a simple Wavefront obj file, preserving only vertices and faces.
func (m *Mesh) WriteObj(filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}

	for _, vert := range m.Verts {
		fmt.Fprintf(w, "v %0.5f %0.5f %0.5f\n", vert.X, vert.Y, vert.Z)
	}

	// Since the face order doesn't matter, sort them to make diffs easier.
	sortedFaces := append([]FaceT{}, m.Faces...)

	// sort first pass - rearrange each face such that the vertex with
	// the smallest index is listed first, keeping the order of the vertices preserved.
	for i, face := range sortedFaces {
		sortedFaces[i] = reorderFace(face)
	}

	sort.Slice(sortedFaces, func(i, j int) bool {
		f1, f2 := sortedFaces[i], sortedFaces[j]
		return cmpFaces(f1, f2)
	})

	for _, face := range sortedFaces {
		indices := make([]string, 0, len(face))
		for _, idx := range face {
			indices = append(indices, fmt.Sprintf("%v", idx+1)) // 1-indexed
		}
		fmt.Fprintf(w, fmt.Sprintf("f %v\n", strings.Join(indices, " ")))
	}

	return w.Close()
}

// reorderFace reorders the vertices such that the smallest index is listed first.
func reorderFace(face FaceT) FaceT {
	f := make(FaceT, 0, len(face))
	start := 0
	smallest := face[0]
	for i := 1; i < len(face); i++ {
		if face[i] < smallest {
			smallest = face[i]
			start = i
		}
	}
	for i := 0; i < len(face); i++ {
		f = append(f, face[(i+start)%len(face)])
	}
	return f
}

// cmpFaces returns true if f1 < f2.
func cmpFaces(f1, f2 FaceT) bool {
	switch {
	case len(f1) == 0:
		return true
	case len(f2) == 0:
		return false
	case f1[0] < f2[0]:
		return true
	case f1[0] > f2[0]:
		return false
	default:
		return cmpFaces(f1[1:], f2[1:])
	}
}
