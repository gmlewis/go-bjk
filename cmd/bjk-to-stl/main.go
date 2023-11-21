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
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/gmlewis/go-bjk/ast"
	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	swapYZ  = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	client := &clientT{c: c}

	for _, arg := range flag.Args() {
		client.processFile(arg)
	}

	log.Printf("Done.")
}

type clientT struct {
	c *nodes.Client
}

func (c *clientT) processFile(arg string) {
	buf, err := os.ReadFile(arg)
	must(err)
	var opts []participle.ParseOption
	if *debug {
		opts = append(opts, participle.Trace(os.Stderr))
	}
	design, err := ast.Parser.ParseString("", string(buf), opts...)
	must(err)
	outFilename := strings.Replace(arg, ".bjk", ".stl", -1)
	log.Printf("Writing STL file: %v", outFilename)
	must(c.c.ToSTL(design, outFilename, *swapYZ))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
