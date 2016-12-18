package gobj

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Vertex represents a 4D vertex.
type Vertex [4]float64

// NewVertex2D returns a 2D Vertex.
func NewVertex2D(x, y float64) Vertex {
	return Vertex{x, y, 0, 0}
}

// NewVertex3D returns a 3D Vertex.
func NewVertex3D(x, y, z float64) Vertex {
	return Vertex{x, y, z, 0}
}

// NewVertex4D returns a 4D Vertex.
func NewVertex4D(x, y, z, w float64) Vertex {
	return Vertex{x, y, z, w}
}

// Scale scales every coordinates by a given scale factor.
func (v *Vertex) Scale(f float64) {
	for i := range v {
		v[i] *= f
	}
}

// X returns the vertex X component.
func (v Vertex) X() float64 {
	return v[0]
}

// Y returns the vertex Y component.
func (v Vertex) Y() float64 {
	return v[1]
}

// Z returns the vertex Z component.
func (v Vertex) Z() float64 {
	return v[2]
}

// W returns the vertex W component.
func (v Vertex) W() float64 {
	return v[3]
}

// Set initializes a vertex from a string array where every string represents a
// vertex component.
func (v *Vertex) Set(s []string) error {
	var (
		err error
	)

	if len(s) > 4 {
		return errors.New("Vertex.Set: invalid string length")
	}

	for i := range s {
		if v[i], err = strconv.ParseFloat(s[i], 64); err != nil {
			return fmt.Errorf("invalid syntax \"%v\": %s", s[i], err)
		}
	}

	return nil
}

// Triangle represents a 3-sided polygon
//
// NOTE: this could easily be extended to support N-sided polygons
// by using a []Vertex instead
type Triangle struct {
	P1, P2, P3 Vertex
}

func (t *Triangle) Scale(f float64) {
	t.P1.Scale(f)
	t.P2.Scale(f)
	t.P3.Scale(f)
}

type AABB struct {
	MinX, MaxX float64
	MinY, MaxY float64
	MinZ, MaxZ float64
}

// NewAABB initializes the bounding box.
//
// The bouding box will be valid after the first call to extend.
func NewAABB() AABB {
	return AABB{
		MinX: math.MaxFloat64,
		MinY: math.MaxFloat64,
		MinZ: math.MaxFloat64,
		MaxX: -math.MaxFloat64,
		MaxY: -math.MaxFloat64,
		MaxZ: -math.MaxFloat64,
	}
}

func (bb *AABB) extend(other AABB) {
	// update the min and max for each coord
	updateMin(&bb.MinX, other.MinX)
	updateMin(&bb.MinY, other.MinY)
	updateMin(&bb.MinZ, other.MinZ)
	updateMax(&bb.MaxX, other.MaxX)
	updateMax(&bb.MaxY, other.MaxY)
	updateMax(&bb.MaxZ, other.MaxZ)
}

func (bb *AABB) Scale(f float64) {
	bb.MinX *= f
	bb.MinY *= f
	bb.MinZ *= f
	bb.MaxX *= f
	bb.MaxY *= f
	bb.MaxZ *= f
}

func (bb AABB) String() string {
	return fmt.Sprintf("x[%f, %f], y[%f, %f], z[%f, %f]",
		bb.MinX, bb.MaxX,
		bb.MinY, bb.MaxY,
		bb.MinZ, bb.MaxZ)
}

// AABB computes and returns the axis-aligned bounding-box
// of the triangle.
func (t Triangle) AABB() AABB {
	return AABB{
		MinX: math.Min(t.P1.X(), math.Min(t.P2.X(), t.P3.X())),
		MinY: math.Min(t.P1.Y(), math.Min(t.P2.Y(), t.P3.Y())),
		MinZ: math.Min(t.P1.Z(), math.Min(t.P2.Z(), t.P3.Z())),
		MaxX: math.Max(t.P1.X(), math.Max(t.P2.X(), t.P3.X())),
		MaxY: math.Max(t.P1.Y(), math.Max(t.P2.Y(), t.P3.Y())),
		MaxZ: math.Max(t.P1.Z(), math.Max(t.P2.Z(), t.P3.Z())),
	}
}

func (t Triangle) isDegenerate() bool {
	// find the determinant of the 3x3 matrix in which the triangle coords can
	// be represented, it's 0 or close to 0.0, the triangle area is null and we
	// consider the triangle as degenerate
	det := (t.P1.X() * t.P2.Y() * t.P3.Z()) +
		(t.P1.Y() * t.P2.Z() * t.P3.X()) +
		(t.P1.Z() * t.P2.X() * t.P3.Y()) -
		(t.P1.Z() * t.P2.Y() * t.P3.X()) -
		(t.P1.Y() * t.P2.X() * t.P3.Z()) -
		(t.P1.X() * t.P2.Z() * t.P3.Y())

	return math.Abs(det) < 1e-5
}

type OBJFile struct {
	vertices  []Vertex
	triangles []Triangle
	aabb      AABB
}

func (of OBJFile) Vertices() []Vertex {
	return of.vertices
}

func (of OBJFile) Triangles() []Triangle {
	return of.triangles
}

func (of OBJFile) AABB() AABB {
	return of.aabb
}

func (of *OBJFile) parseVertex(kw string, data []string) error {
	v := Vertex{}
	err := v.Set(data)
	if err != nil {
		return err
	}
	// discard the Z coordinate
	of.vertices = append(of.vertices, v)
	return nil
}

func (of *OBJFile) parseFace(kw string, data []string) error {
	if len(data) != 3 {
		return errors.New("only triangular faces are supported")
	}

	var (
		vi  [3]int
		err error
	)

	// vertices indices
	for i, s := range data {
		// we are only interested in the vertex index
		sidx := strings.Split(s, "/")[0]
		vi[i], err = strconv.Atoi(sidx)
		if err != nil {
			return fmt.Errorf("invalid vertex coordinate value \"%s\"", s)
		}
	}

	t := Triangle{
		P1: of.vertices[vi[0]-1],
		P2: of.vertices[vi[1]-1],
		P3: of.vertices[vi[2]-1],
	}

	// extend the mesh bounding box with the triangle
	of.aabb.extend(t.AABB())

	of.triangles = append(of.triangles, t)
	return nil
}

func (of *OBJFile) DumpInfo() string {
	var nfo string
	nfo = fmt.Sprintln("num vertices :", len(of.vertices))
	nfo += fmt.Sprintln("num triangles:", len(of.triangles))
	nfo += fmt.Sprintln("bounding box :", of.aabb)
	return nfo
}

// updateMin checks if a > b, then a will be set to the value of b.
func updateMin(a *float64, b float64) {
	if b < *a {
		*a = b
	}
}

// updateMax checks if a < b, then a will be set to the value of b.
func updateMax(a *float64, b float64) {
	if *a < b {
		*a = b
	}
}