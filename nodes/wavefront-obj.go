package nodes

import (
	"fmt"
	"os"
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
	for _, face := range m.Faces {
		indices := make([]string, 0, len(face))
		for _, idx := range face {
			indices = append(indices, fmt.Sprintf("%v", idx+1)) // 1-indexed
		}
		fmt.Fprintf(w, "f %v\n", strings.Join(indices, " "))
	}

	return w.Close()
}
