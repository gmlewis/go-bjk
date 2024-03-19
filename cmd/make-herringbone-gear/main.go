// -*- compile-command: "go run main.go"; -*-

// make-herringbone-gear tests the STL output for HerringboneGear.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug      = flag.Bool("debug", false, "Turn on debugging info")
	helixAngle = flag.Float64("ha", 45, "Helix angle")
	objOut     = flag.String("obj", "make-herringbone-gear.obj", "Output filename for Wavefront obj file")
	outBJK     = flag.String("o", "make-herringbone-gear.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir    = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	stlOut     = flag.String("stl", "make-herringbone-gear.stl", "Output filename for binary STL file")
	swapYZ     = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	r := 22.5
	for i := 0; i < 26; i++ {
		ratio := float64(i) / 26
		x := r * math.Cos(2*math.Pi*ratio)
		y := r * math.Sin(2*math.Pi*ratio)
		log.Printf("i=%v, angle=%.2f=%0.2f, (%.2f, %0.2f, 30)", i, 2*math.Pi*ratio, 360*ratio, x, y)
	}

	design, err := c.NewBuilder().AddNode("HerringboneGear", set("helix_angle", *helixAngle)).Build()
	must(err)

	if *outBJK == "-" {
		fmt.Printf("%v\n", design)
	} else if *outBJK != "" {
		must(os.WriteFile(*outBJK, []byte(design.String()+"\n"), 0644))
	}

	if *objOut != "" {
		must(c.ToObj(design, *objOut))
	}

	if *stlOut != "" {
		must(c.ToSTL(design, *stlOut, *swapYZ))
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
