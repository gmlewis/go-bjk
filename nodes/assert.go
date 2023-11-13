package nodes

import "log"

func assert(v bool, errMsg string) {
	if !v {
		log.Fatal(errMsg)
	}
}
