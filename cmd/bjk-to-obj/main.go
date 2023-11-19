// bjk-to-obj loads a Blackjack BJK file and writes a Wavefront Obj file.
// See: https://github.com/setzer22/blackjack
//
// Usage:
//
//	bjk-to-obj file.bjk [file2.bjk ...]
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/gmlewis/go-bjk/ast"
	"github.com/gmlewis/go-bjk/nodes"
	"github.com/mitchellh/go-homedir"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
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

	for _, arg := range flag.Args() {
		buf, err := os.ReadFile(arg)
		must(err)
		design, err := ast.Parser.ParseString("", string(buf), participle.Trace(os.Stderr))
		must(err)
		outFilename := strings.Replace(arg, ".bjk", ".obj", -1)
		log.Printf("Writing Wavefront obj file: %v", outFilename)
		must(c.ToObj(design, outFilename))
	}

	log.Printf("Done.")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
