package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"gradient"
)

func linearGradient(x0, y0, x1, y1 float64, fname string) error {
	width, height := 512, 512
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	stops := []gradient.Stop{{0.0, color.NRGBA{255, 0, 0, 255}},
		{0.5, color.NRGBA{0, 255, 0, 16}},
		{1.0, color.NRGBA{0, 0, 255, 255}}}

	gradient.DrawLinear(img, x0, y0, x1, y1, stops)

	var (
		err error
		f   *os.File
	)

	if f, err = os.Create(fname); err != nil {
		return err
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		return err
	}

	return nil
}

func radialGradient(cx, cy, r, fx, fy float64, fname string) error {
	width, height := 512, 512
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	stops := []gradient.Stop{{0.0, color.NRGBA{255, 0, 0, 255}},
		{0.5, color.NRGBA{0, 255, 0, 16}},
		{1.0, color.NRGBA{0, 0, 255, 255}}}

	gradient.DrawRadial(img, cx, cy, r, fx, fy, stops)

	var (
		err error
		f   *os.File
	)

	if f, err = os.Create(fname); err != nil {
		return err
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		return err
	}

	return nil
}

func main() {
	var err error

	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err = linearGradient(0.1, 0.1, 0.9, 0.1, "hlinear.png"); err != nil {
		return
	}
	if err = linearGradient(0.1, 0.1, 0.1, 0.9, "vlinear.png"); err != nil {
		return
	}
	if err = linearGradient(0.1, 0.1, 0.9, 0.9, "linear.png"); err != nil {
		return
	}
	if err = radialGradient(0.5, 0.5, 0.5, 0.5, 0.5, "simpleradial.png"); err != nil {
		return
	}
	if err = radialGradient(0.5, 0.5, 0.5, 0.7, 0.7, "focusradial.png"); err != nil {
		return
	}
}
