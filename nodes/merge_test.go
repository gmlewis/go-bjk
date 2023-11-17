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

	loadObj := func(t *testing.T, filename string) *Mesh {
		t.Helper()
		buf, err := goldenObjs.ReadFile(filepath.Join("testdata", filename))
		if err != nil {
			t.Fatal(err)
		}

		m, err := ObjStrToMesh(string(buf))
		if err != nil {
			t.Fatal(err)
		}
		return m
	}

	for _, prefix := range testCasePrefixes {
		t.Run(prefix, func(t *testing.T) {
			src := loadObj(t, prefix+"-src.obj")
			dst := loadObj(t, prefix+"-dst.obj")
			want := loadObj(t, prefix+"-result.obj")
			log.Printf("merging src '%v' into dst '%v'", prefix+"-src.obj", prefix+"-dst.obj")
			t.Logf("merging src '%v' into dst '%v'", prefix+"-src.obj", prefix+"-dst.obj")
			dst.Merge(src)
			compareMeshes(t, prefix+"-result.obj", dst, want)

			src = loadObj(t, prefix+"-src.obj")
			dst = loadObj(t, prefix+"-dst.obj")
			want = loadObj(t, prefix+"-swapped-result.obj")
			log.Printf("merging dst '%v' into src '%v'", prefix+"-dst.obj", prefix+"-src.obj")
			t.Logf("merging dst '%v' into src '%v'", prefix+"-dst.obj", prefix+"-src.obj")
			src.Merge(dst)
			compareMeshes(t, prefix+"-swapped-result.obj", src, want)
		})
	}
}

func compareMeshes(t *testing.T, name string, got, want *Mesh) {
	t.Helper()

	if len(got.uniqueVerts) != len(want.uniqueVerts) {
		t.Fatalf("%v: got %v uniqueVerts, want %v", name, len(got.uniqueVerts), len(want.uniqueVerts))
	}

	if len(got.Faces) != len(want.Faces) {
		t.Fatalf("%v: got %v faces, want %v", name, len(got.Faces), len(want.Faces))
	}
}
