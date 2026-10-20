// -*- compile-command: "go run main.go -nt 12"; -*-

// make-planetary-gearbox generates a print-in-place planetary gearbox using
// the Blackjack HerringboneGear node. The center "sun" gear, three orbiting
// "planet" gears, and the enclosing inward-facing "ring" gear are all
// HerringboneGear nodes. The ring gear uses gear_side=Internal, which generates
// an internal herringbone profile as the envelope swept by the mating planet
// gear (see herringbone_gear.lua in the Blackjack repo).
//
// Gear terminology and formulas follow the HerringboneGear node:
//
//	pitch radius = module * num_teeth / 2
//	base radius  = pitch radius * cos(pressure_angle)
//	root radius  = pitch radius - 1.25 * module
//	outer radius = pitch radius + module
//
// For the internal/ring gear the addendum and dedendum are reversed:
//
//	ring tooth tip radius (inward) = ring pitch radius - 1.25 * module
//	ring root/notch radius (outer) = ring pitch radius + module
//
// For a planetary set with a sun, Np planet gears at 120 degrees, and an
// internal ring gear, the tooth counts must satisfy:
//
//	ring = sun + 2*planet
//	(sun + ring) % planets == 0
//
// With all four inner gears identical (planet == sun), the second condition
// becomes sun % 3 == 0 for three planets. The inherited default of 13 teeth
// therefore cannot be equally spaced; use -nt 12, -nt 15, or -planetTeeth 11
// (or 14).
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/gmlewis/go-bjk/nodes"
)

const deg = math.Pi / 180.0

var (
	debug          = flag.Bool("debug", false, "Turn on debugging info")
	gearLength     = flag.Float64("gl", 30, "Gear length")
	helixAngle     = flag.Float64("ha", 30, "Helix angle")
	holeRadius     = flag.Float64("hr", 10, "Hole minimum radius")
	holeType       = flag.String("ht", "None", "Hole type (one of: 'None', 'Hollow', 'Squared', 'Hexagonal', 'Octagonal', 'Circular')")
	module         = flag.Float64("mod", 3, "Gear module - see https://www.stlgears.com/theory#module")
	numElbows      = flag.Int("ne", 4, "Number of elbows")
	numTeeth       = flag.Int("nt", 13, "Number of teeth on the center (sun) gear")
	numPlanets     = flag.Int("planets", 3, "Number of orbiting planet gears")
	planetTeeth    = flag.Int("planetTeeth", 0, "Number of teeth on each planet gear (0 means same as -nt)")
	pressureAngle  = flag.Float64("pa", 20, "Pressure angle in degrees")
	ringTeeth      = flag.Int("ringTeeth", 0, "Number of teeth on the outer ring gear (0 means sun + 2*planet)")
	tolerance      = flag.Float64("tol", 0.25, "Target minimum 3D-printer clearance in millimeters between meshing tooth flanks")
	outerThickness = flag.Float64("outerThickness", 10.0, "Radial wall thickness from the closest ring notch to the outer smooth surface, in millimeters")
	objOut         = flag.String("obj", "make-planetary-gearbox.obj", "Output filename for Wavefront obj file")
	outBJK         = flag.String("o", "make-planetary-gearbox.bjk", "Output filename for BJK file ('-' for stdout, '' for none)")
	repoDir        = flag.String("repo", "src/github.com/gmlewis/blackjack", "Path to Blackjack repo (relative to home dir or absolute path)")
	resolution     = flag.Int("res", 9, "Resolution for number of points in curved sections")
	stlOut         = flag.String("stl", "make-planetary-gearbox.stl", "Output filename for binary STL file")
	swapYZ         = flag.Bool("swapyz", true, "Swap Y and Z values when writing STL file (Wavefront obj always swaps for Blender)")
)

func main() {
	flag.Parse()

	cfg, err := computeConfig()
	must(err)

	c, err := nodes.New(*repoDir, *debug)
	must(err)
	defer c.Close()

	// Build the four inner gears and the internal ring gear.
	b := c.NewBuilder()
	b = addGear(b, "sun", cfg.sun, vec3{X: 0, Y: 0, Z: 0}, cfg.phiSunDeg, cfg.sunDirection, cfg.helixAngleSun)
	for j, p := range cfg.planets {
		name := fmt.Sprintf("planet-%v", j)
		b = addGear(b, name, cfg.planet, p.pos, p.phiDeg, cfg.planetDirection, cfg.helixAngleDeg)
	}
	b = addRingGear(b, cfg)

	// Merge everything into the final mesh. The last node is the default node.
	b = b.MergeMesh("HerringboneGear.ring.out_mesh")
	b = b.MergeMesh("HerringboneGear.sun.out_mesh")
	for j := range cfg.planets {
		b = b.MergeMesh(fmt.Sprintf("HerringboneGear.planet-%v.out_mesh", j))
	}
	design, err := b.Build()
	must(err)

	if *outBJK == "-" {
		fmt.Printf("%v\n", design)
	} else if *outBJK != "" {
		must(os.WriteFile(*outBJK, []byte(design.String()+"\n"), 0644))
		fmt.Printf("Wrote %v\n", *outBJK)
	}

	// Evaluate each gear separately and concatenate the meshes. The parts do
	// not intersect (print-in-place clearance), so Blackjack's face-cutting
	// MergeMeshes pass is unnecessary for the STL/OBJ output and is avoided
	// here to keep the output geometry exactly as each node generated it.
	var meshes []*nodes.Mesh
	meshes = append(meshes, evalRingGear(c, cfg))
	meshes = append(meshes, evalGear(c, "sun", cfg.sun, vec3{X: 0, Y: 0, Z: 0}, cfg.phiSunDeg, cfg.sunDirection, cfg.helixAngleSun))
	for j, p := range cfg.planets {
		meshes = append(meshes, evalGear(c, fmt.Sprintf("planet-%v", j), cfg.planet, p.pos, p.phiDeg, cfg.planetDirection, cfg.helixAngleDeg))
	}
	fullMesh := combineMeshes(meshes...)

	if *objOut != "" {
		must(fullMesh.WriteObj(*objOut))
		fmt.Printf("Wrote %v\n", *objOut)
	}

	if *stlOut != "" {
		must(fullMesh.WriteSTL(*stlOut, *swapYZ))
		fmt.Printf("Wrote %v\n", *stlOut)
	}

	printConfig(cfg)

	log.Printf("Done.")
}

func evalGear(c *nodes.Client, name string, g gearDims, pos vec3, phiDeg float64, direction string, helix float64) *nodes.Mesh {
	b := c.NewBuilder()
	b = addGear(b, name, g, pos, phiDeg, direction, helix)
	design, err := b.Build()
	must(err)
	mesh, err := c.Eval(design)
	must(err)
	return mesh
}

func evalRingGear(c *nodes.Client, cfg config) *nodes.Mesh {
	b := c.NewBuilder()
	b = addRingGear(b, cfg)
	design, err := b.Build()
	must(err)
	mesh, err := c.Eval(design)
	must(err)
	return mesh
}

// combineMeshes concatenates the faces of the meshes into a new mesh.
func combineMeshes(meshes ...*nodes.Mesh) *nodes.Mesh {
	out := nodes.NewMesh()
	for _, m := range meshes {
		if m == nil {
			continue
		}
		for _, face := range m.Faces {
			verts := make([]nodes.Vec3, len(face))
			for i, idx := range face {
				verts[i] = m.Verts[idx]
			}
			out.AddFace(verts)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

type gearDims struct {
	module      float64
	numTeeth    int
	pitchRadius float64
	baseRadius  float64
	rootRadius  float64
	outerRadius float64
}

type vec3 struct {
	X, Y, Z float64
}

type planetConfig struct {
	pos    vec3
	phiDeg float64
}

type config struct {
	sun       gearDims
	planet    gearDims
	ring      gearDims
	ringTeeth int

	gearLength     float64
	numElbows      int
	resolution     int
	helixAngleDeg  float64
	helixAngleSun  float64 // generator value that matches the planets at the pitch circle
	pressureAngle  float64
	holeType       string
	holeRadius     float64
	tolerance      float64
	outerThickness float64

	centerDistance float64 // sun/planet center distance (ring/planet distance is the same)
	centerDelta    float64 // increase over the zero-backlash center distance

	phiSunDeg       float64
	sunDirection    string
	planetDirection string
	planets         []planetConfig
}

func computeConfig() (config, error) {
	var cfg config

	switch {
	case *module <= 0:
		return cfg, fmt.Errorf("-mod must be > 0")
	case *gearLength <= 0:
		return cfg, fmt.Errorf("-gl must be > 0")
	case *numTeeth < 6:
		return cfg, fmt.Errorf("-nt must be >= 6 (Blackjack's HerringboneGear minimum)")
	case *numElbows < 1:
		return cfg, fmt.Errorf("-ne must be >= 1")
	case *resolution < 1:
		return cfg, fmt.Errorf("-res must be >= 1")
	case *numPlanets < 1:
		return cfg, fmt.Errorf("-planets must be >= 1")
	case *pressureAngle < 1:
		return cfg, fmt.Errorf("-pa must be >= 1")
	case *tolerance < 0:
		return cfg, fmt.Errorf("-tol must be >= 0")
	case *outerThickness <= 0:
		return cfg, fmt.Errorf("-outerThickness must be > 0")
	}

	pTeeth := *planetTeeth
	if pTeeth == 0 {
		pTeeth = *numTeeth
	}
	if pTeeth < 6 {
		return cfg, fmt.Errorf("-planetTeeth must be >= 6 (or 0 to use -nt)")
	}

	rTeeth := *ringTeeth
	if rTeeth == 0 {
		rTeeth = *numTeeth + 2*pTeeth
	}
	if rTeeth != *numTeeth+2*pTeeth {
		return cfg, fmt.Errorf("-ringTeeth (%v) must equal -nt + 2*(-planetTeeth) = %v + 2*%v = %v for a standard internal mesh",
			rTeeth, *numTeeth, pTeeth, *numTeeth+2*pTeeth)
	}

	// Assembly condition for equally-spaced planets.
	if (*numTeeth+rTeeth)%*numPlanets != 0 {
		return cfg, fmt.Errorf("tooth counts (sun=%v, planet=%v, ring=%v) cannot be equally spaced in %v positions: "+
			"(sun+ring)=%v must be divisible by %v.\n"+
			"  With all four inner gears identical, this requires -nt to be divisible by 3.\n"+
			"  Try: -nt 12, -nt 15, or -planetTeeth %v (ring=%v)",
			*numTeeth, pTeeth, rTeeth, *numPlanets,
			*numTeeth+rTeeth, *numPlanets,
			suggestPlanetTeeth(*numTeeth, *numPlanets), *numTeeth+2*suggestPlanetTeeth(*numTeeth, *numPlanets))
	}

	pa := *pressureAngle * deg
	ha := math.Max(0, math.Min(45, *helixAngle)) * deg
	cfg.gearLength = *gearLength
	cfg.numElbows = *numElbows
	cfg.resolution = *resolution
	cfg.helixAngleDeg = ha / deg // clamped to the HerringboneGear node's range
	cfg.pressureAngle = *pressureAngle
	cfg.holeType = *holeType
	cfg.holeRadius = *holeRadius
	cfg.tolerance = *tolerance
	cfg.outerThickness = *outerThickness

	cfg.sun = newGearDims(*module, pa, *numTeeth)
	cfg.planet = newGearDims(*module, pa, pTeeth)
	cfg.ringTeeth = rTeeth
	cfg.ring = newGearDims(*module, pa, rTeeth)
	// Nominal internal-gear radii for reporting.
	cfg.ring.rootRadius = cfg.ring.pitchRadius + *module       // deepest notch (outer)
	cfg.ring.outerRadius = cfg.ring.pitchRadius - 1.25**module // inward-facing tooth tip

	// Increase the sun/planet center distance so the actual normal flank
	// clearance is the requested tolerance. The same center distance is passed
	// to the ring node, whose envelope is generated at that distance.
	cfg.centerDelta = 0
	if *tolerance > 0 {
		cfg.centerDelta = *tolerance * 2
	}
	cfg.centerDistance = cfg.sun.pitchRadius + cfg.planet.pitchRadius + cfg.centerDelta

	// Match the helix angle at the pitch circle. The Lua node defines the
	// helix angle at the outer radius, so the pitch-circle angle is
	// tan(beta_pitch) = tan(ha) * N/(N+2).
	tanBetaPitch := math.Tan(ha) * float64(pTeeth) / float64(pTeeth+2)
	cfg.helixAngleSun = math.Atan(tanBetaPitch*float64(*numTeeth+2)/float64(*numTeeth)) / deg
	cfg.helixAngleSun = math.Max(0, math.Min(45, cfg.helixAngleSun))

	// Sun phase is 0: tooth 0 of the sun points along +X.
	cfg.phiSunDeg = 0
	cfg.sunDirection = "Counter-Clockwise" // direction = +1
	cfg.planetDirection = "Clockwise"      // direction = -1 (opposite to the sun for an external mesh)

	// Place the planets and compute their rotations.
	cfg.planets = make([]planetConfig, *numPlanets)
	for j := 0; j < *numPlanets; j++ {
		theta := 2 * math.Pi * float64(j) / float64(*numPlanets)
		s := frac(float64(*numTeeth) * theta / (2 * math.Pi))
		// External sun/planet mesh: frac_sun(theta) + frac_planet(theta+pi) = 1/2.
		phi := theta + math.Pi - (2*math.Pi/float64(pTeeth))*(0.5-s)
		cfg.planets[j] = planetConfig{
			pos: vec3{
				X: cfg.centerDistance * math.Cos(theta),
				Y: 0,
				Z: cfg.centerDistance * math.Sin(theta),
			},
			phiDeg: normAngle(phi) / deg,
		}
	}

	return cfg, nil
}

func newGearDims(module, pressureAngle float64, numTeeth int) gearDims {
	pitch := module * float64(numTeeth) / 2
	return gearDims{
		module:      module,
		numTeeth:    numTeeth,
		pitchRadius: pitch,
		baseRadius:  pitch * math.Cos(pressureAngle),
		rootRadius:  pitch - 1.25*module,
		outerRadius: pitch + module,
	}
}

// suggestPlanetTeeth returns a planet tooth count close to sunTeeth that makes
// (sun + ring) = 2*(sun + planet) divisible by the number of planets.
func suggestPlanetTeeth(sunTeeth, planets int) int {
	for offset := 1; offset < 20; offset++ {
		for _, p := range []int{sunTeeth - offset, sunTeeth + offset} {
			if p < 6 {
				continue
			}
			ring := sunTeeth + 2*p
			if (sunTeeth+ring)%planets == 0 {
				return p
			}
		}
	}
	return sunTeeth
}

// ---------------------------------------------------------------------------
// Blackjack design
// ---------------------------------------------------------------------------

func addGear(b *nodes.Builder, name string, g gearDims, pos vec3, phiDeg float64, direction string, helix float64) *nodes.Builder {
	return b.AddNode("HerringboneGear."+name,
		set("pos", fmt.Sprintf("vector(%v,%v,%v)", pos.X, pos.Y, pos.Z)),
		set("speed", 0),
		set("speed_units", "TeethPerSecond"),
		set("pivot", "Center"),
		set("pivot_rotation", 0),
		set("gear_rotation", phiDeg),
		set("module", g.module),
		set("direction", direction),
		set("num_teeth", g.numTeeth),
		set("num_elbows", *numElbows),
		set("helix_angle", helix),
		set("resolution", *resolution),
		set("pressure_angle", *pressureAngle),
		set("gear_length", *gearLength),
		set("hole_type", *holeType),
		set("hole_radius", *holeRadius),
	)
}

func addRingGear(b *nodes.Builder, cfg config) *nodes.Builder {
	return b.AddNode("HerringboneGear.ring",
		set("pos", "vector(0,0,0)"),
		set("speed", 0),
		set("speed_units", "TeethPerSecond"),
		set("pivot", "Center"),
		set("pivot_rotation", 0),
		set("gear_rotation", 0),
		set("module", cfg.ring.module),
		set("direction", cfg.planetDirection),
		set("num_teeth", cfg.ringTeeth),
		set("num_elbows", cfg.numElbows),
		set("helix_angle", cfg.helixAngleDeg),
		set("resolution", cfg.resolution),
		set("pressure_angle", cfg.pressureAngle),
		set("gear_length", cfg.gearLength),
		set("hole_type", "None"),
		set("hole_radius", 0),
		set("gear_side", "Internal"),
		set("mate_num_teeth", cfg.planet.numTeeth),
		set("mate_rotation", cfg.planets[0].phiDeg),
		set("center_distance", cfg.centerDistance),
		set("clearance", cfg.tolerance),
		set("outer_thickness", cfg.outerThickness),
	)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func frac(x float64) float64 {
	return x - math.Floor(x)
}

func normAngle(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}

func printConfig(cfg config) {
	fmt.Printf("sun:    teeth=%v pitch_radius=%v base_radius=%v root_radius=%v outer_radius=%v rotation=%v deg\n",
		cfg.sun.numTeeth, cfg.sun.pitchRadius, cfg.sun.baseRadius, cfg.sun.rootRadius, cfg.sun.outerRadius, cfg.phiSunDeg)
	fmt.Printf("planet: teeth=%v pitch_radius=%v base_radius=%v root_radius=%v outer_radius=%v\n",
		cfg.planet.numTeeth, cfg.planet.pitchRadius, cfg.planet.baseRadius, cfg.planet.rootRadius, cfg.planet.outerRadius)
	fmt.Printf("ring:   teeth=%v pitch_radius=%v base_radius=%v nominal_tooth_tip=%v nominal_notch=%v\n",
		cfg.ring.numTeeth, cfg.ring.pitchRadius, cfg.ring.baseRadius, cfg.ring.outerRadius, cfg.ring.rootRadius)
	fmt.Printf("center distance (sun-planet and ring-planet): %v (zero-backlash center distance + %v)\n",
		cfg.centerDistance, cfg.centerDelta)
	fmt.Printf("target minimum clearance: %v, outer thickness: %v\n", cfg.tolerance, cfg.outerThickness)
	for j, p := range cfg.planets {
		fmt.Printf("planet %v: pos=(%0.6f, 0, %0.6f) rotation=%0.6f deg\n", j, p.pos.X, p.pos.Z, p.phiDeg)
	}
	fmt.Printf("helix angle: %v deg (sun generator: %v deg)\n", cfg.helixAngleDeg, cfg.helixAngleSun)
}

func set(key string, value any) string {
	return fmt.Sprintf("%v=%v", key, value)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
