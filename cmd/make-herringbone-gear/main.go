// -*- compile-command: "go run main.go"; -*-

// make-herringbone-gear tests the STL output for HerringboneGear.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

var (
	debug      = flag.Bool("debug", false, "Turn on debugging info")
	helixAngle = flag.Float64("ha", 30, "Helix angle")
	holeRadius = flag.Float64("hr", 10, "Hole minimum radius")
	holeType   = flag.String("ht", "None", "Hole type (one of: 'None', 'Hollow', 'Squared', 'Hexagonal', 'Octagonal', 'Circular')")
	module     = flag.Float64("mod", 3, "Gear module - see https://www.stlgears.com/theory#module")
	numTeeth   = flag.Int("nt", 13, "Number of teeth")
	objOut     = flag.String("obj", "make-herringbone-gear.obj", "Output filename for Wavefront obj file")
	outBJK     = flag.String("o", "make-herringbone-gear.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir    = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	resolution = flag.Int("res", 9, "Resolution for number of points in curved sections")
	stlOut     = flag.String("stl", "make-herringbone-gear.stl", "Output filename for binary STL file")
	swapYZ     = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	design, err := c.NewBuilder().AddNode(
		"HerringboneGear",
		set("helix_angle", *helixAngle),
		set("hole_radius", *holeRadius),
		set("hole_type", *holeType),
		set("module", *module),
		set("num_teeth", *numTeeth),
		set("resolution", *resolution),
	).Build()
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
