// -*- compile-command: "go build cone.go && ./cone && fstl cone.stl"; -*-

package main

import (
	"log"

	. "github.com/gmlewis/pt/pt"
)

const (
	baseRadius = 0.5
	topRadius  = 0.0
	height     = 1.0
	bbox       = 2.0
	step       = 0.01
)

func main() {
	cone := NewConeSDF(baseRadius, topRadius, height, 0)
	mesh := NewSDFMesh(cone, Box{Vector{-bbox, -bbox, -bbox}, Vector{bbox, bbox, bbox}}, step)
	if err := SaveSTL("cone.stl", mesh); err != nil {
		log.Fatalf("Unable to write cone.stl: %v", err)
	}
}
