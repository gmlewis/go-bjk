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
	gearLength = flag.Float64("gl", 30, "Gear length")
	helixAngle = flag.Float64("ha", 30, "Helix angle")
	holeRadius = flag.Float64("hr", 10, "Hole minimum radius")
	holeType   = flag.String("ht", "None", "Hole type (one of: 'None', 'Hollow', 'Squared', 'Hexagonal', 'Octagonal', 'Circular')")
	module     = flag.Float64("mod", 3, "Gear module - see https://www.stlgears.com/theory#module")
	numElbows  = flag.Int("ne", 1, "Number of elbows")
	numTeeth   = flag.Int("nt", 13, "Number of teeth")
	objOut     = flag.String("obj", "make-herringbone-gear.obj", "Output filename for Wavefront obj file")
	outBJK     = flag.String("o", "make-herringbone-gear.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	pivot      = flag.String("pivot", "PitchRadius", "Sets the gear pivot point to one of: 'Center', 'RootRadius', 'BaseRadius', 'PitchRadius', or 'OuterRadius'.")
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
		set("gear_length", *gearLength),
		set("helix_angle", *helixAngle),
		set("hole_radius", *holeRadius),
		set("hole_type", *holeType),
		set("module", *module),
		set("num_elbows", *numElbows),
		set("num_teeth", *numTeeth),
		set("pivot", *pivot),
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

	br, err := c.GetScalar(design, "HerringboneGear.base_radius")
	must(err)
	fmt.Printf("base_radius: %v\n", br)
	pr, err := c.GetScalar(design, "HerringboneGear.pitch_radius")
	must(err)
	fmt.Printf("pitch_radius: %v\n", pr)
	or, err := c.GetScalar(design, "HerringboneGear.outer_radius")
	must(err)
	fmt.Printf("outer_radius: %v\n", or)
	rr, err := c.GetScalar(design, "HerringboneGear.root_radius")
	must(err)
	fmt.Printf("root_radius: %v\n", rr)

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
