package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
)

type blockResult struct {
	blockX    int
	blockY    int
	Rectangle [blockSize][blockSize]uint8
}

const (
	fname     string     = "mandelbrot.png"
	pStart    complex128 = (-2.0 + -1.5i)
	pEnd      complex128 = (0.6 + 1.5i)
	points    int        = 2000
	maxIters  int        = 256
	blockSize int        = 32
)

var (
	Version string
	Build   string
)

func compute(bx int, by int, out chan blockResult) {
	xStep := (real(pEnd) - real(pStart)) / float64(points)
	yStep := (imag(pEnd) - imag(pStart)) / float64(points)

	var ret blockResult

	ret.blockX = bx
	ret.blockY = by
	for x := 0; x < blockSize; x++ {
		for y := 0; y < blockSize; y++ {
			cReal := real(pStart) + float64(x+blockSize*bx)*xStep
			cImag := imag(pStart) + float64(y+blockSize*by)*yStep
			c := complex(cReal, cImag)
			z := complex(0, 0)
			curIters := maxIters
			for i := 1; i < maxIters; i++ {
				z = cmplx.Pow(z, 2) + c
				if real(z)+imag(z) > 4 {
					curIters = i
					break
				}
			}
			ret.Rectangle[x][y] = uint8(curIters)
		}
	}

	out <- ret
}

func main() {
	fmt.Printf("Starting mandelbrot: version=%s build=%s\n", Version, Build)

	img := image.NewGray(image.Rect(0, 0, points, points))
	results := make(chan blockResult)
	var res blockResult

	for i := 0; i < int(points/blockSize); i++ {
		for j := 0; j < int(points/blockSize); j++ {
			go compute(i, j, results)
		}
	}

	for i := 0; i < int(points/blockSize); i++ {
		for j := 0; j < int(points/blockSize); j++ {
			res = <-results
			for x, ycol := range res.Rectangle {
				for y, r := range ycol {
					img.Set(x+blockSize*res.blockX, y+blockSize*res.blockY, color.Gray{r})
				}
			}
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
