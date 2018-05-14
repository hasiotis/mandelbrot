package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"sync"

	pb "github.com/hasiotis/mandelbrot/v5/rpc"
	"github.com/mediocregopher/radix.v2/pool"
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
	p       *pool.Pool
	mux     sync.Mutex
)

func getCachedBlock(i int, j int) ([blockSize][blockSize]uint8, bool) {
	var cached bool = false
	var unserialized [blockSize][blockSize]uint8
	blockid := fmt.Sprintf("%03d%03d", i, j)
	if p != nil {
		mux.Lock()
		exists, err := p.Cmd("HEXISTS", "mandel", blockid).Int()
		if err != nil {
			log.Fatal(err)
		}
		if exists == 1 {
			v, err := p.Cmd("HGET", "mandel", blockid).Bytes()
			if err != nil {
				log.Fatal(err)
			} else {
				err := json.Unmarshal(v, &unserialized)
				if err != nil {
					log.Printf("Failed cache unmarshal: error=%s", err)
				} else {
					cached = true
				}
			}
		}
		mux.Unlock()
	}
	return unserialized, cached
}

func setCachedBlock(i int, j int, r [blockSize][blockSize]uint8) {
	blockid := fmt.Sprintf("%03d%03d", i, j)
	if p != nil {
		serialized, err := json.Marshal(r)
		if err != nil {
			log.Printf("Serialize failed [%s]: blockid=%s\n", err, blockid)
		}
		mux.Lock()

		reply, err := p.Cmd("HSET", "mandel", blockid, serialized).Int()
		if err != nil {
			log.Fatal(err)
		} else if reply == 1 {
			log.Printf("Updated: blockid=%s\n", blockid)
		}
		mux.Unlock()
	}
}

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
				var cached bool
				var ret blockResult
				ret.blockX = i
				ret.blockY = j
				ret.Rectangle, cached = getCachedBlock(i, j)
				if !cached {
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
					setCachedBlock(i, j, ret.Rectangle)
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
	var err error
	p, err = pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		log.Printf("Redis server is not reachable: error=%s\n", err)
	}

	img := calculateMandel()
	sendImage(w, img)
}

func main() {
	log.Printf("Starting mandelbrot: version=%s build=%s\n", Version, Build)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
