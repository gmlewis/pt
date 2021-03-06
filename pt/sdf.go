package pt

import "math"

// SDFShape

type SDFShape struct {
	SDF
	Material Material
}

func NewSDFShape(sdf SDF, material Material) Shape {
	return &SDFShape{sdf, material}
}

func (s *SDFShape) Compile() {
}

func (s *SDFShape) Intersect(ray Ray) Hit {
	const epsilon = 0.00001
	const start = 0.0001
	const jumpSize = 0.001
	box := s.BoundingBox()
	t1, t2 := box.Intersect(ray)
	if t2 < t1 || t2 < 0 {
		return NoHit
	}
	t := math.Max(start, t1)
	jump := true
	for i := 0; i < 1000; i++ {
		d := s.Evaluate(ray.Position(t))
		if jump && d < 0 {
			t -= jumpSize
			jump = false
			continue
		}
		if d < epsilon {
			return Hit{s, t, nil}
		}
		if jump && d < jumpSize {
			d = jumpSize
		}
		t += d
		if t > t2 {
			return NoHit
		}
	}
	return NoHit
}

func (s *SDFShape) UV(p Vector) Vector {
	return Vector{}
}

func (s *SDFShape) NormalAt(p Vector) Vector {
	const e = 0.0001
	x, y, z := p.X, p.Y, p.Z
	n := Vector{
		s.Evaluate(Vector{x - e, y, z}) - s.Evaluate(Vector{x + e, y, z}),
		s.Evaluate(Vector{x, y - e, z}) - s.Evaluate(Vector{x, y + e, z}),
		s.Evaluate(Vector{x, y, z - e}) - s.Evaluate(Vector{x, y, z + e}),
	}
	return n.Normalize()
}

func (s *SDFShape) MaterialAt(p Vector) Material {
	return s.Material
}

// SDF

type SDF interface {
	Evaluate(p Vector) float64
	BoundingBox() Box
}

// SphereSDF

type SphereSDF struct {
	Radius   float64
	Exponent float64
}

func NewSphereSDF(radius float64) SDF {
	return &SphereSDF{radius, 2}
}

func (s *SphereSDF) Evaluate(p Vector) float64 {
	return p.LengthN(s.Exponent) - s.Radius
}

func (s *SphereSDF) BoundingBox() Box {
	r := s.Radius
	return Box{Vector{-r, -r, -r}, Vector{r, r, r}}
}

// V2

type V2 struct {
	X, Y float64
}

func (v V2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v V2) Sub(d V2) V2 {
	return V2{v.X - d.X, v.Y - d.Y}
}

func (v V2) Dot(d V2) float64 {
	return v.X*d.X + v.Y*d.Y
}

func (v V2) Normalize() V2 {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y)
	if length > 0 {
		return V2{v.X / length, v.Y / length}
	}
	return v
}

// ConeSDF

type ConeSDF struct {
	BaseRadius float64
	TopRadius  float64
	Height     float64
	Round      float64
}

func NewConeSDF(baseRadius, topRadius, height, round float64) SDF {
	return &ConeSDF{baseRadius, topRadius, height, round}
}

func (s *ConeSDF) Evaluate(p Vector) float64 {
	// Cone algorithn from https://github.com/deadsy/sdfx
	height := s.Height/2 - s.Round
	// cone slope vector and normal
	su := V2{s.TopRadius, height / 2}.Sub(V2{s.BaseRadius, -height / 2}).Normalize()
	sn := V2{su.Y, -su.X}
	// inset the radii for the rounding
	ofs := s.Round / sn.X
	sr0 := s.BaseRadius - (1+sn.Y)*ofs
	sr1 := s.TopRadius - (1-sn.Y)*ofs
	// cone slope length
	sl := V2{sr1, height}.Sub(V2{sr0, -height}).Length()

	// convert to SoR 2d coordinates
	p2 := V2{V2{p.X, p.Y}.Length(), p.Z}
	// is p2 above the cone?
	if p2.Y >= height && p2.X <= sr1 {
		return p2.Y - height - s.Round
	}
	// is p2 below the cone?
	if p2.Y <= -height && p2.X <= sr0 {
		return -p2.Y - height - s.Round
	}
	// distance to slope line
	v := p2.Sub(V2{sr0, -height})
	d_slope := v.Dot(sn)
	// is p2 inside the cone?
	if d_slope < 0 && math.Abs(p2.Y) < height {
		return -math.Min(-d_slope, height-math.Abs(p2.Y)) - s.Round
	}
	// is p2 closest to the slope line?
	t := v.Dot(su)
	if t >= 0 && t <= sl {
		return d_slope - s.Round
	}
	// is p2 closest to the base radius vertex?
	if t < 0 {
		return v.Length() - s.Round
	}
	// p2 is closest to the top radius vertex
	return p2.Sub(V2{sr1, height}).Length() - s.Round
}

func (s *ConeSDF) BoundingBox() Box {
	r := s.BaseRadius
	if s.TopRadius > r {
		r = s.TopRadius
	}
	h := s.Height / 2
	return Box{Vector{-r, -h, -r}, Vector{r, h, r}}
}

// CubeSDF

type CubeSDF struct {
	Size Vector
}

func NewCubeSDF(size Vector) SDF {
	return &CubeSDF{size}
}

func (s *CubeSDF) Evaluate(p Vector) float64 {
	x := p.X
	y := p.Y
	z := p.Z
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	if z < 0 {
		z = -z
	}
	x -= s.Size.X / 2
	y -= s.Size.Y / 2
	z -= s.Size.Z / 2
	a := x
	if y > a {
		a = y
	}
	if z > a {
		a = z
	}
	if a > 0 {
		a = 0
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if z < 0 {
		z = 0
	}
	b := math.Sqrt(x*x + y*y + z*z)
	return a + b
}

func (s *CubeSDF) BoundingBox() Box {
	x, y, z := s.Size.X/2, s.Size.Y/2, s.Size.Z/2
	return Box{Vector{-x, -y, -z}, Vector{x, y, z}}
}

// CylinderSDF

type CylinderSDF struct {
	Radius float64
	Height float64
}

func NewCylinderSDF(radius, height float64) SDF {
	return &CylinderSDF{radius, height}
}

func (s *CylinderSDF) Evaluate(p Vector) float64 {
	x := math.Sqrt(p.X*p.X + p.Z*p.Z)
	y := p.Y
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	x -= s.Radius
	y -= s.Height / 2
	a := x
	if y > a {
		a = y
	}
	if a > 0 {
		a = 0
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	b := math.Sqrt(x*x + y*y)
	return a + b
}

func (s *CylinderSDF) BoundingBox() Box {
	r := s.Radius
	h := s.Height / 2
	return Box{Vector{-r, -h, -r}, Vector{r, h, r}}
}

// CapsuleSDF

type CapsuleSDF struct {
	A, B     Vector
	Radius   float64
	Exponent float64
}

func NewCapsuleSDF(a, b Vector, radius float64) SDF {
	return &CapsuleSDF{a, b, radius, 2}
}

func (s *CapsuleSDF) Evaluate(p Vector) float64 {
	pa := p.Sub(s.A)
	ba := s.B.Sub(s.A)
	h := math.Max(0, math.Min(1, pa.Dot(ba)/ba.Dot(ba)))
	return pa.Sub(ba.MulScalar(h)).LengthN(s.Exponent) - s.Radius
}

func (s *CapsuleSDF) BoundingBox() Box {
	a, b := s.A.Min(s.B), s.A.Max(s.B)
	return Box{a.SubScalar(s.Radius), b.AddScalar(s.Radius)}
}

// TorusSDF

type TorusSDF struct {
	MajorRadius   float64
	MinorRadius   float64
	MajorExponent float64
	MinorExponent float64
}

func NewTorusSDF(major, minor float64) SDF {
	return &TorusSDF{major, minor, 2, 2}
}

func (s *TorusSDF) Evaluate(p Vector) float64 {
	q := Vector{Vector{p.X, p.Y, 0}.LengthN(s.MajorExponent) - s.MajorRadius, p.Z, 0}
	return q.LengthN(s.MinorExponent) - s.MinorRadius
}

func (s *TorusSDF) BoundingBox() Box {
	a := s.MinorRadius
	b := s.MinorRadius + s.MajorRadius
	return Box{Vector{-b, -b, -a}, Vector{b, b, a}}
}

// TransformSDF

type TransformSDF struct {
	SDF
	Matrix  Matrix
	Inverse Matrix
}

func NewTransformSDF(sdf SDF, matrix Matrix) SDF {
	return &TransformSDF{sdf, matrix, matrix.Inverse()}
}

func (s *TransformSDF) Evaluate(p Vector) float64 {
	q := s.Inverse.MulPosition(p)
	return s.SDF.Evaluate(q)
}

func (s *TransformSDF) BoundingBox() Box {
	return s.Matrix.MulBox(s.SDF.BoundingBox())
}

// ScaleSDF

type ScaleSDF struct {
	SDF
	Factor float64
}

func NewScaleSDF(sdf SDF, factor float64) SDF {
	return &ScaleSDF{sdf, factor}
}

func (s *ScaleSDF) Evaluate(p Vector) float64 {
	return s.SDF.Evaluate(p.DivScalar(s.Factor)) * s.Factor
}

func (s *ScaleSDF) BoundingBox() Box {
	f := s.Factor
	m := Scale(Vector{f, f, f})
	return m.MulBox(s.SDF.BoundingBox())
}

// UnionSDF

type UnionSDF struct {
	Items []SDF
}

func NewUnionSDF(items ...SDF) SDF {
	return &UnionSDF{items}
}

func (s *UnionSDF) Evaluate(p Vector) float64 {
	var result float64
	for i, item := range s.Items {
		d := item.Evaluate(p)
		if i == 0 || d < result {
			result = d
		}
	}
	return result
}

func (s *UnionSDF) BoundingBox() Box {
	var result Box
	for i, item := range s.Items {
		box := item.BoundingBox()
		if i == 0 {
			result = box
		} else {
			result = result.Extend(box)
		}
	}
	return result
}

// DifferenceSDF

type DifferenceSDF struct {
	Items []SDF
}

func NewDifferenceSDF(items ...SDF) SDF {
	return &DifferenceSDF{items}
}

func (s *DifferenceSDF) Evaluate(p Vector) float64 {
	var result float64
	for i, item := range s.Items {
		d := item.Evaluate(p)
		if i == 0 {
			result = d
		} else if -d > result {
			result = -d
		}
	}
	return result
}

func (s *DifferenceSDF) BoundingBox() Box {
	return s.Items[0].BoundingBox()
}

// IntersectionSDF

type IntersectionSDF struct {
	Items []SDF
}

func NewIntersectionSDF(items ...SDF) SDF {
	return &IntersectionSDF{items}
}

func (s *IntersectionSDF) Evaluate(p Vector) float64 {
	var result float64
	for i, item := range s.Items {
		d := item.Evaluate(p)
		if i == 0 || d > result {
			result = d
		}
	}
	return result
}

func (s *IntersectionSDF) BoundingBox() Box {
	// TODO: intersect boxes
	var result Box
	for i, item := range s.Items {
		box := item.BoundingBox()
		if i == 0 {
			result = box
		} else {
			result = result.Extend(box)
		}
	}
	return result
}

// RepeatSDF

type RepeatSDF struct {
	SDF
	Step Vector
}

func NewRepeatSDF(sdf SDF, step Vector) SDF {
	return &RepeatSDF{sdf, step}
}

func (s *RepeatSDF) Evaluate(p Vector) float64 {
	q := p.Mod(s.Step).Sub(s.Step.DivScalar(2))
	return s.SDF.Evaluate(q)
}

func (s *RepeatSDF) BoundingBox() Box {
	// TODO: fix this
	return Box{}
}
