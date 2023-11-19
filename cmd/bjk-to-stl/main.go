// bjk-to-stl loads a Blackjack BJK file and writes an STL file.
// See: https://github.com/setzer22/blackjack
//
// Usage:
//
//	bjk-to-stl file.bjk [file2.bjk ...]
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
	swapYZ  = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file")
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
		outFilename := strings.Replace(arg, ".bjk", ".stl", -1)
		log.Printf("Writing STL file: %v", outFilename)
		must(c.ToSTL(design, outFilename, *swapYZ))
	}

	log.Printf("Done.")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
