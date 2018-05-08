package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"strconv"

	pb "github.com/hasiotis/mandelbrot/v4/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

func calculateMandel() *image.Gray {
	img := image.NewGray(image.Rect(0, 0, points, points))

	conn, err := grpc.Dial("localhost:28000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMandelServiceClient(conn)

	results := make(chan blockResult)
	var res blockResult

	for i := 0; i < int(points/blockSize); i++ {
		for j := 0; j < int(points/blockSize); j++ {
			go func(i int, j int) {
				var ret blockResult
				ret.blockX = i
				ret.blockY = j
				ps := &pb.ComplexPoint{real(pStart), imag(pStart)}
				pe := &pb.ComplexPoint{real(pEnd), imag(pEnd)}
				r, err := c.ComputeMandel(context.Background(), &pb.BlockRequest{ps, pe, int32(points), int32(maxIters), int32(blockSize), int32(i), int32(j)})
				if err != nil {
					log.Fatalf("could not request compute: %v", err)
				}
				for x := 0; x < blockSize; x++ {
					for y := 0; y < blockSize; y++ {
						ret.Rectangle[x][y] = uint8(r.Results[x*blockSize+y])
					}
				}
				results <- ret
			}(i, j)
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

	return img
}

func sendImage(w http.ResponseWriter, img *image.Gray) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	img := calculateMandel()
	sendImage(w, img)
}

func main() {
	log.Printf("Staring mandelbrot: version=%s build=%s\n", Version, Build)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
