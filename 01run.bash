#!/bin/bash

set -e
SCRIPT_DIR="${0%/*}"

cd "$SCRIPT_DIR"/cmd/video-ocr/
go run main.go "$@"
