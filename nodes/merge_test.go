package nodes

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/golden*obj
var goldenObjs embed.FS

func TestMerge(t *testing.T) {
	t.Parallel()

	dirEntries, err := goldenObjs.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	var testCasePrefixes []string
	for _, de := range dirEntries {
		filename := de.Name()
		if strings.HasSuffix(filename, "-src.obj") {
			testCasePrefixes = append(testCasePrefixes, strings.TrimSuffix(filename, "-src.obj"))
		}
	}

	tempDir, err := os.MkdirTemp("", "golden-tests")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // clean up

	for _, prefix := range testCasePrefixes {
		if prefix != "golden-3-edge-cut" {
			continue // debug only
		}

		t.Run(prefix, func(t *testing.T) {
			{ // These blocks make it easy to disable one or the other with a leading "if false".
				src := loadObj(t, prefix+"-src.obj")
				dst := loadObj(t, prefix+"-dst.obj")
				// log.Printf("merging src '%v' into dst '%v'", prefix+"-src.obj", prefix+"-dst.obj")
				// t.Logf("merging src '%v' into dst '%v'", prefix+"-src.obj", prefix+"-dst.obj")
				dst.Merge(src)
				want, err := maybeLoadObj(t, prefix+"-result.obj")
				if err != nil {
					t.Error(err)
					log.Printf("writing %v-got-result.obj", prefix)
					dst.WriteObj(prefix + "-got-result.obj")
					return
				}
				compareMeshes(t, prefix+"-result.obj", dst, want)
			}

			{
				src := loadObj(t, prefix+"-src.obj")
				dst := loadObj(t, prefix+"-dst.obj")
				// log.Printf("merging dst '%v' into src '%v'", prefix+"-dst.obj", prefix+"-src.obj")
				// t.Logf("merging dst '%v' into src '%v'", prefix+"-dst.obj", prefix+"-src.obj")
				src.Merge(dst)
				want, err := maybeLoadObj(t, prefix+"-swapped-result.obj")
				if err != nil {
					t.Error(err)
					log.Printf("writing %v-got-swapped-result.obj", prefix)
					src.WriteObj(prefix + "-got-swapped-result.obj")
					return
				}
				compareMeshes(t, prefix+"-swapped-result.obj", src, want)
			}
		})
	}
}

func loadObj(t *testing.T, filename string) *Mesh {
	t.Helper()
	m, err := maybeLoadObj(t, filename)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func maybeLoadObj(t *testing.T, filename string) (*Mesh, error) {
	buf, err := goldenObjs.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		return nil, err
	}

	m, err := ObjStrToMesh(string(buf))
	if err != nil {
		return nil, err
	}
	return m, nil
}

func compareMeshes(t *testing.T, name string, got, want *Mesh) {
	t.Helper()

	if len(got.uniqueVerts) != len(want.uniqueVerts) {
		t.Errorf("%v: got %v uniqueVerts, want %v", name, len(got.uniqueVerts), len(want.uniqueVerts))
	}

	if len(got.Faces) != len(want.Faces) {
		t.Errorf("%v: got %v faces, want %v", name, len(got.Faces), len(want.Faces))
	}
}
