// -*- compile-command: "go build cyl.go && ./cyl && fstl cyl.stl"; -*-

package main

import (
	"log"

	. "github.com/gmlewis/pt/pt"
)

const (
	radius = 1.0
	height = 1.0
	bbox   = 2.0
	step   = 0.01
)

func main() {
	cyl := NewCylinderSDF(radius, height)
	mesh := NewSDFMesh(cyl, Box{Vector{-bbox, -bbox, -bbox}, Vector{bbox, bbox, bbox}}, step)
	if err := SaveSTL("cyl.stl", mesh); err != nil {
		log.Fatalf("Unable to write cyl.stl: %v", err)
	}
}
