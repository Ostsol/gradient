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

// DrawHLinear draws a linear gradient to dst, and is optimized for producing a
// purely horizontal gradient.
func DrawHLinear(dst draw.Image, x0, x1 float64, stops []Stop) {
	bb := dst.Bounds()
	width := bb.Dx()

	x0, x1 = x0*float64(width), x1*float64(width)
	dx := x1 - x0

	var col color.Color

	for x := 0; x < width; x++ {
		fx := float64(x)
		if fx < x0 {
			col = stops[0].Col
		} else {
			d := (fx - x0) / dx
			col = stops[len(stops)-1].Col
			for i, stop := range stops[1:] {
				if d < stop.X {
					d = (d - stops[i].X) / (stop.X - stops[i].X)
					col = collerp(stops[i].Col, stop.Col, d)
					break
				}
			}
		}

		for y := bb.Min.Y; y < bb.Max.Y; y++ {
			dst.Set(x+bb.Min.X, y+bb.Min.Y, col)
		}
	}
}

// DrawVLinear draws a linear gradient to dst, and is optimized for producing a
// purely vertical gradient.
func DrawVLinear(dst draw.Image, y0, y1 float64, stops []Stop) {
	bb := dst.Bounds()
	height := bb.Dy()

	y0, y1 = y0*float64(height), y1*float64(height)
	dy := y1 - y0

	var col color.Color

	for y := 0; y < height; y++ {
		fy := float64(y)
		if fy < y0 {
			col = stops[0].Col
		} else {
			d := (fy - y0) / dy
			col = stops[len(stops)-1].Col
			for i, stop := range stops[1:] {
				if d < stop.X {
					d = (d - stops[i].X) / (stop.X - stops[i].X)
					col = collerp(stops[i].Col, stop.Col, d)
					break
				}
			}
		}

		for x := bb.Min.X; x < bb.Max.X; x++ {
			dst.Set(x, y+bb.Min.Y, col)
		}
	}
}

// DrawLinear draws a linear gradient to dst. If the gradient vector (as
// defined by x0, y0, x1, and y1) is found to be purely horizontal or purely
// vertical, the appropriate optimized functions will be called.
func DrawLinear(dst draw.Image, x0, y0, x1, y1 float64, stops []Stop) {
	if x0 > x1 {
		panic(fmt.Sprintf("invalid bounds x0(%f)>x1(%f)", x0, x1))
	}

	if y0 == y1 && x0 != x1 {
		DrawHLinear(dst, x0, x1, stops)
		return
	}

	if y0 > y1 {
		panic(fmt.Sprintf("invalid bounds y0(%f)>y1(%f)", y0, y1))
	}

	if x0 == x1 && y0 != y1 {
		DrawVLinear(dst, y0, y1, stops)
		return
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

				if d < stops[0].X {
					col = stops[0].Col
				} else {
					col = stops[len(stops)-1].Col
					// iterate through stops to find the colour range
					for i, st := range stops[1:] {
						if d < st.X {
							d = (d - stops[i].X) / (st.X - stops[i].X)
							col = collerp(stops[i].Col, st.Col, d)
							break
						}
					}
				}
			}
			dst.Set(x+bb.Min.X, y+bb.Min.Y, col)
		}
	}
}
