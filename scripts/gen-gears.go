// -*- compile-command: "go run gen-gears.go"; -*-

// Usage:
// go run gen-gears.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
)

var (
	minTeeth = flag.Int("min", 6, "Minimum number of teeth")
	maxTeeth = flag.Int("max", 9, "Minimum number of teeth")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	done := map[int]bool{}
	for i := range *maxTeeth - *minTeeth + 1 {
		teeth := *minTeeth + i
		if done[teeth] {
			continue
		}
		makeGear(teeth)
		done[teeth] = true
		for j := range *maxTeeth - *minTeeth + 1 {
			t2 := teeth * (*minTeeth + j)
			if done[t2] {
				continue
			}
			makeGear(t2)
			done[t2] = true
		}
	}

	log.Printf("Done.")
}

func makeGear(teeth int) {
	stlName := fmt.Sprintf("gear%v-20.stl", teeth)
	log.Printf("Generating '%v'...", stlName)
	args := []string{"./herringbone-gear.sh", "-nt", fmt.Sprintf("%v", teeth), "-gl", "20", "-stl", stlName}
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	fmt.Printf("%s\n", out)
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/*

go run gen-gears.go
2024/04/14 16:29:42 Generating 'gear6-20.stl'...
base_radius: 8.457233587073176
pitch_radius: 9
outer_radius: 12
root_radius: 5.25
2024/04/14 16:29:43 Done.

2024/04/14 16:29:43 Generating 'gear36-20.stl'...
base_radius: 50.74340152243906
pitch_radius: 54
outer_radius: 57
root_radius: 50.25
2024/04/14 16:29:48 Done.

2024/04/14 16:29:48 Generating 'gear42-20.stl'...
base_radius: 59.200635109512234
pitch_radius: 63
outer_radius: 66
root_radius: 59.25
2024/04/14 16:29:53 Done.

2024/04/14 16:29:53 Generating 'gear48-20.stl'...
base_radius: 67.65786869658541
pitch_radius: 72
outer_radius: 75
root_radius: 68.25
2024/04/14 16:29:59 Done.

2024/04/14 16:29:59 Generating 'gear54-20.stl'...
base_radius: 76.11510228365859
pitch_radius: 81
outer_radius: 84
root_radius: 77.25
2024/04/14 16:30:06 Done.

2024/04/14 16:30:06 Generating 'gear7-20.stl'...
base_radius: 9.866772518252038
pitch_radius: 10.5
outer_radius: 13.5
root_radius: 6.75
2024/04/14 16:30:07 Done.

2024/04/14 16:30:07 Generating 'gear49-20.stl'...
base_radius: 69.06740762776427
pitch_radius: 73.5
outer_radius: 76.5
root_radius: 69.75
2024/04/14 16:30:13 Done.

2024/04/14 16:30:13 Generating 'gear56-20.stl'...
base_radius: 78.9341801460163
pitch_radius: 84
outer_radius: 87
root_radius: 80.25
2024/04/14 16:30:21 Done.

2024/04/14 16:30:21 Generating 'gear63-20.stl'...
base_radius: 88.80095266426835
pitch_radius: 94.5
outer_radius: 97.5
root_radius: 90.75
2024/04/14 16:30:29 Done.

2024/04/14 16:30:29 Generating 'gear8-20.stl'...
base_radius: 11.276311449430901
pitch_radius: 12
outer_radius: 15
root_radius: 8.25
2024/04/14 16:30:30 Done.

2024/04/14 16:30:30 Generating 'gear64-20.stl'...
base_radius: 90.21049159544721
pitch_radius: 96
outer_radius: 99
root_radius: 92.25
2024/04/14 16:30:38 Done.

2024/04/14 16:30:38 Generating 'gear72-20.stl'...
base_radius: 101.48680304487812
pitch_radius: 108
outer_radius: 111
root_radius: 104.25
2024/04/14 16:30:47 Done.

2024/04/14 16:30:47 Generating 'gear9-20.stl'...
base_radius: 12.685850380609764
pitch_radius: 13.5
outer_radius: 16.5
root_radius: 9.75
2024/04/14 16:30:49 Done.

2024/04/14 16:30:49 Generating 'gear81-20.stl'...
base_radius: 114.17265342548788
pitch_radius: 121.5
outer_radius: 124.5
root_radius: 117.75
2024/04/14 16:30:59 Done.

*/
