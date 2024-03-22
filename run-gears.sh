#!/bin/bash -ex
go run cmd/make-herringbone-gear/main.go -ht None -stl test-gear.stl -obj test-gear.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Hollow -stl test-gear-hollow.stl -obj test-gear-hollow.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Squared -stl test-gear-squared.stl -obj test-gear-squared.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Hexagonal -stl test-gear-hexagonal.stl -obj test-gear-hexagonal.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Octagonal -stl test-gear-octagonal.stl -obj test-gear-octagonal.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Circular -stl test-gear-circular.stl -obj test-gear-circular.obj -o '' "$@"
go run cmd/make-herringbone-gear/main.go -ht Octagonal -stl test-gear-octagonal-center.stl -obj test-gear-octagonal-center.obj -o '' -pivot Center "$@"

fstl test-gear.stl
fstl test-gear-hollow.stl
fstl test-gear-squared.stl
fstl test-gear-hexagonal.stl
fstl test-gear-octagonal.stl
fstl test-gear-circular.stl
fstl test-gear-octagonal-center.stl
