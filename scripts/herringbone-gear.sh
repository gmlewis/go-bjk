#!/bin/bash -e

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
REPO_DIR=$(realpath ${SCRIPT_DIR}/..)

go run ${REPO_DIR}/cmd/make-herringbone-gear/main.go -ne 3 -obj '' -o '' "$@"
