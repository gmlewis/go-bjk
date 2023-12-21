#!/bin/bash -ex
go run cmd/extrude-helix/main.go "$@"
go run cmd/extrude-quad/main.go "$@"
go run cmd/make-bfem-cage/main.go "$@"
go run cmd/make-box/main.go "$@"
go run cmd/make-elbows/main.go "$@"
go run cmd/make-svgpath/main.go "$@"
go run examples/bifilar-electromagnet/main.go "$@"
