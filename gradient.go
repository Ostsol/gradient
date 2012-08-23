// Copyright 2012 Daniel Jo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gradient draws gradients with an api similar to that specified by
// SVG.
package gradient

import (
	"fmt"
	"image/color"
	"image/draw"
	"math"
)

// Stop is a gradient control point, where Col is the colour at X.
type Stop struct {
	// X is normally in the range [0,1], but values outside the range are
	// also accepted.
	X   float64
	Col color.Color
}

// lerp is a linear interpolation between two unsigned 32-bit integers, return
// in an unsigned 8-bit integer for use in an RGBA colour.
func lerp(a, b uint32, x float64) uint8 {
	return uint8(int32(float64(a)*(1.0-x)+float64(b)*x) >> 8)
}

// collerp perfoms a linear interpolation between two colours.
func collerp(c0, c1 color.Color, x float64) color.Color {
	r0, g0, b0, a0 := c0.RGBA()
	r1, g1, b1, a1 := c1.RGBA()
	return color.NRGBA{lerp(r0, r1, x),
		lerp(g0, g1, x),
		lerp(b0, b1, x),
		lerp(a0, a1, x)}
}

// getColour iterates through the gradient stops to find the stop in which the
// pixel resides and returns an interpolated colour.

func getColour(rat float64, stops []Stop) color.Color {
	if rat <= 0.0 || len(stops) == 1 {
		return stops[0].Col
	}

	last := stops[len(stops)-1]

	if rat >= last.X {
		return last.Col
	}

	for i, stop := range stops[1:] {
		if rat < stop.X {
			rat = (rat - stops[i].X) / (stop.X - stops[i].X)
			return collerp(stops[i].Col, stop.Col, rat)
		}
	}

	return last.Col
}

// drawHLinear draws a linear gradient to dst, and is optimized for producing a
// purely horizontal gradient.
func drawHLinear(dst draw.Image, x0, x1 float64, stops []Stop) {
	if x0 > x1 {
		panic(fmt.Sprintf("invalid bounds x0(%f)>x1(%f)", x0, x1))
	}

	if len(stops) == 0 {
		return
	}

	bb := dst.Bounds()
	width := bb.Dx()

	x0, x1 = x0*float64(width), x1*float64(width)
	dx := x1 - x0

	for x := 0; x < width; x++ {
		col := getColour((float64(x)-x0)/dx, stops)

		for y := bb.Min.Y; y < bb.Max.Y; y++ {
			dst.Set(x+bb.Min.X, y+bb.Min.Y, col)
		}
	}
}

// drawVLinear draws a linear gradient to dst, and is optimized for producing a
// purely vertical gradient.
func drawVLinear(dst draw.Image, y0, y1 float64, stops []Stop) {
	if y0 > y1 {
		panic(fmt.Sprintf("invalid bounds y0(%f)>y1(%f)", y0, y1))
	}

	if len(stops) == 0 {
		return
	}

	bb := dst.Bounds()
	height := bb.Dy()

	y0, y1 = y0*float64(height), y1*float64(height)
	dy := y1 - y0

	for y := 0; y < height; y++ {
		col := getColour((float64(y)-y0)/dy, stops)

		for x := bb.Min.X; x < bb.Max.X; x++ {
			dst.Set(x, y+bb.Min.Y, col)
		}
	}
}

// DrawLinear draws a linear gradient to dst. If the gradient vector (as
// defined by x0, y0, x1, and y1) is found to be purely horizontal or purely
// vertical, the appropriate optimized functions will be called.
func DrawLinear(dst draw.Image, x0, y0, x1, y1 float64, stops []Stop) {
	if y0 == y1 && x0 != x1 {
		drawHLinear(dst, x0, x1, stops)
		return
	}

	if x0 == x1 && y0 != y1 {
		drawVLinear(dst, y0, y1, stops)
		return
	}

	if len(stops) == 0 {
		return
	}

	if y0 > y1 {
		panic(fmt.Sprintf("invalid bounds y0(%f)>y1(%f)", y0, y1))
	}
	if x0 > x1 {
		panic(fmt.Sprintf("invalid bounds x0(%f)>x1(%f)", x0, x1))
	}

	bb := dst.Bounds()
	width, height := bb.Dx(), bb.Dy()

	x0, y0 = x0*float64(width), y0*float64(height)
	x1, y1 = x1*float64(width), y1*float64(height)

	dx, dy := x1-x0, y1-y0
	px0, py0 := x0-dy, y0+dx
	mag := math.Hypot(dx, dy)

	var col color.Color
	for y := 0; y < width; y++ {
		fy := float64(y)

		for x := 0; x < width; x++ {
			fx := float64(x)
			// is the pixel before the start of the gradient?
			s0 := (px0-x0)*(fy-y0) - (py0-y0)*(fx-x0)
			if s0 > 0 {
				col = stops[0].Col
			} else {
				// calculate the distance of the pixel from the first stop line
				u := ((fx-x0)*(px0-x0) + (fy-y0)*(py0-y0)) /
					(mag * mag)
				x2, y2 := x0+u*(px0-x0), y0+u*(py0-y0)
				d := math.Hypot(fx-x2, fy-y2) / mag

				col = getColour(d, stops)
			}
			dst.Set(x+bb.Min.X, y+bb.Min.Y, col)
		}
	}
}

// drawSimpleRadial draws a simplified radial gradient with a centred focus.
// It represents the gradient as a right cone with radius r and height 1.0. The
// cone equation is x^2/a^2 + y^2/b^2 = z^2/c^2. Solving for z gives me the
// ratio with which to interpolate step colours.

func drawSimpleRadial(dst draw.Image, cx, cy, r float64, stops []Stop) {
	if len(stops) == 0 {
		return
	}

	bb := dst.Bounds()
	width, height := bb.Dx(), bb.Dy()

	a, b := r*float64(width), r*float64(height)
	cx, cy = cx*float64(width), cy*float64(height)

	for x := 0; x < width; x++ {
		x2_a2 := ((float64(x) - cx) * (float64(x) - cx)) / (a * a)
		for y := 0; y < height; y++ {
			y2 := (float64(y) - cy) * (float64(y) - cy)
			rat := math.Sqrt(x2_a2 + y2/(b*b))

			dst.Set(x+bb.Min.X, y+bb.Min.Y, getColour(rat, stops))
		}
	}
}

// DrawRadial draws a radial gradient centred at cx, cy, with radius r, focused
// at fx, fy, into dst. All numerical values are a treated as fraction of the
// relevant dimension of dst. Values outside of [0.0,1.0] are accepted.
// A quick path is used when the focus is at 0.0,0.0.
// The algorithm is adapted from Maxim Shemanarev's Anti-Grain Geometry,
// http://www.antigrain.com. Relevant file: agg_span_gradient.h.

func DrawRadial(dst draw.Image, cx, cy, r, fx, fy float64, stops []Stop) {
	if len(stops) == 0 {
		return
	}

	if fx == cx && fy == cy {
		drawSimpleRadial(dst, cx, cy, r, stops)
		return
	}

	bb := dst.Bounds()
	width, height := bb.Dx(), bb.Dy()

	fx = (fx - cx) * float64(width)
	fy = (fy - cy) * float64(height)
	cx *= float64(width)
	cy *= float64(height)
	r *= float64(width)

	yrat := float64(height) / float64(width)

	f := math.Hypot(fx, fy)
	if f > r {
		fx = fx / f * (r - 1)
		fy = fy / f * (r - 1)
	}

	r2 := r * r
	mul := r / (r2 - (fx*fx + fy*fy))

	for x := 0; x < width; x++ {
		dx := float64(x) - cx - fx
		dx2 := dx * dx
		for y := 0; y < height; y++ {
			dy := (float64(y) - cy - fy) / yrat
			d2 := dx*fy - dy*fx
			d3 := r2*(dx2+dy*dy) - d2*d2
			rat := (dx*fx + dy*fy + math.Sqrt(math.Abs(d3))) * mul / r
			dst.Set(x+bb.Min.X, y+bb.Min.Y, getColour(rat, stops))
		}
	}
}
