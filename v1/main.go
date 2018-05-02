package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
)

const (
	fname    string     = "mandelbrot.png"
	pStart   complex128 = (-2.0 + -1.5i)
	pEnd     complex128 = (0.6 + 1.5i)
	points   int        = 2000
	maxIters int        = 256
)

var (
	Version string
	Build   string
)

func main() {
	fmt.Printf("Starting mandelbrot: version=%s build=%s\n", Version, Build)

	img := image.NewGray(image.Rect(0, 0, points, points))
	xStep := (real(pEnd) - real(pStart)) / float64(points)
	yStep := (imag(pEnd) - imag(pStart)) / float64(points)
	for x := 0; x < points; x++ {
		for y := 0; y < points; y++ {
			cReal := real(pStart) + float64(x)*xStep
			cImag := imag(pStart) + float64(y)*yStep
			c := complex(cReal, cImag)
			z := complex(0, 0)
			iters := 0
			for i := 1; i < maxIters; i++ {
				z = cmplx.Pow(z, 2) + c
				if real(z)+imag(z) > 4 {
					iters = i
					break
				}
			}
			img.Set(x, y, color.Gray{uint8(iters)})
		}
	}

	f, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		fmt.Println(err)
	}

	if err := f.Close(); err != nil {
		fmt.Println(err)
	}
}
