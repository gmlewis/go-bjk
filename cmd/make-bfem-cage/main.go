// -*- compile-command: "go run main.go -stl ../../make-bfem-cage.stl"; -*-

// make-bfem-cage tests the STL output for BFEMCage.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug    = flag.Bool("debug", false, "Turn on debugging info")
	repoDir  = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	segments = flag.Int("ns", 36, "Number of segments")
	stlOut   = flag.String("stl", "make-bfem-cage.stl", "Output filename for binary STL file")
)

func main() {
	flag.Parse()

	homeDir, err := homedir.Dir()
	must(err)

	repoPath := filepath.Join(homeDir, *repoDir)
	c, err := nodes.New(repoPath, *debug)
	if err != nil {
		c, err = nodes.New(*repoDir, *debug)
		must(err)
	}
	defer c.Close()

	design, err := c.NewBuilder().AddNode("BFEMCage", set("segments", *segments)).Build()
	must(err)

	if *stlOut != "" {
		must(c.ToSTL(design, *stlOut))
	}

	log.Printf("Done.")
}

func set(key string, value any) string {
	return fmt.Sprintf("%v=%v", key, value)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}