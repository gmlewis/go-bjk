// -*- compile-command: "go run main.go ../../make-box.bjk"; -*-

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
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/gmlewis/go-bjk/ast"
	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug   = flag.Bool("debug", false, "Turn on debugging info")
	repoDir = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	outFile = flag.String("o", "", "Override output filename")
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
	if err != nil {
		log.Fatalf("ERROR: ast.Parser.ParseString: %v", err)
	}
	outFilename := strings.Replace(arg, ".bjk", ".obj", -1)
	if *outFile != "" {
		outFilename = *outFile
	}
	log.Printf("Writing Wavefront obj file: %v", outFilename)
	must(c.c.ToObj(design, outFilename))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
