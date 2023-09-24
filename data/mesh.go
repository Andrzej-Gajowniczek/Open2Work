package main

import (
	"fmt"
	"math"
	"time"
)

const (
	width  = 80
	height = 24
)

func main() {
	// Create a new 3D cube with the specified size
	cube := NewCube(10)

	// Set up the rotation angles
	rotationX := 0.0
	rotationY := 0.0
	rotationZ := 0.0

	for {
		// Clear the console screen
		fmt.Print("\033[2J")
		fmt.Print("\033[H")

		// Rotate the cube
		rotationX += 0.02
		rotationY += 0.03
		rotationZ += 0.01

		// Get the rotated vertices of the cube
		rotatedVertices := cube.Rotate(rotationX, rotationY, rotationZ)

		// Draw the cube on the console
		DrawCube(rotatedVertices)

		// Pause for a short duration before rendering the next frame
		time.Sleep(50 * time.Millisecond)
	}
}

// Vertex represents a vertex in 3D space
type Vertex struct {
	X, Y, Z float64
}

// Cube represents a 3D cube
type Cube struct {
	Size     float64
	Vertices [8]Vertex
}

// NewCube creates a new 3D cube with the specified size
func NewCube(size float64) *Cube {
	halfSize := size / 2

	return &Cube{
		Size: size,
		Vertices: [8]Vertex{
			{-halfSize, -halfSize, -halfSize},
			{-halfSize, -halfSize, halfSize},
			{-halfSize, halfSize, -halfSize},
			{-halfSize, halfSize, halfSize},
			{halfSize, -halfSize, -halfSize},
			{halfSize, -halfSize, halfSize},
			{halfSize, halfSize, -halfSize},
			{halfSize, halfSize, halfSize},
		},
	}
}

// Rotate applies rotation transformations to the cube and returns the rotated vertices
func (c *Cube) Rotate(rotationX, rotationY, rotationZ float64) []Vertex {
	rotatedVertices := make([]Vertex, len(c.Vertices))

	// Apply rotation transformations to each vertex
	for i, vertex := range c.Vertices {
		x, y, z := vertex.X, vertex.Y, vertex.Z

		// Rotate along the x-axis
		y, z = rotate2D(y, z, rotationX)

		// Rotate along the y-axis
		x, z = rotate2D(x, z, rotationY)

		// Rotate along the z-axis
		x, y = rotate2D(x, y, rotationZ)

		rotatedVertices[i] = Vertex{x, y, z}
	}

	return rotatedVertices
}

// rotate2D applies a 2D rotation transformation to the given coordinates
func rotate2D(x, y, angle float64) (float64, float64) {
	radians := angle * math.Pi / 180
	cos := math.Cos(radians)
	sin := math.Sin(radians)

	newX := x*cos - y*sin
	newY := x*sin + y*cos

	return newX, newY
}

// DrawCube draws the cube on the console
func DrawCube(vertices []Vertex) {
	// Define the edges of the cube
	edges := [][]int{
		{0, 1}, {1, 3}, {3, 2}, {2, 0},
		{4, 5}, {5, 7}, {7, 6}, {6, 4},
		{0, 4}, {1, 5}, {3, 7}, {2, 6},
	}

	// Draw the cube edges
	for _, edge := range edges {
		v1 := vertices[edge[0]]
		v2 := vertices[edge[1]]

		drawLine(int(v1.X), int(v1.Y), int(v2.X), int(v2.Y))
	}
}

// drawLine draws a line between two points on the console
func drawLine(x1, y1, x2, y2 int) {
	dx := x2 - x1
	dy := y2 - y1

	absDx := int(math.Abs(float64(dx)))
	absDy := int(math.Abs(float64(dy)))

	if absDx == 0 && absDy == 0 {
		return
	}

	var sx, sy int
	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}

	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}

	err := absDx - absDy

	for {
		fmt.Printf("\033[%d;%dH#", y1+height/2, x1+width/2)

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -absDy {
			err -= absDy
			x1 += sx
		}
		if e2 < absDx {
			err += absDx
			y1 += sy
		}
	}
}

