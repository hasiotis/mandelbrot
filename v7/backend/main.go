package main

import (
	"log"
	"net"

	"math/cmplx"

	pb "github.com/hasiotis/mandelbrot/v7/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	Version string
	Build   string
)

type server struct{}

func (s *server) ComputeMandel(ctx context.Context, in *pb.BlockRequest) (*pb.BlockReply, error) {
	br := new(pb.BlockReply)

	xStep := (in.PEnd.X - in.PStart.X) / float64(in.Points)
	yStep := (in.PEnd.Y - in.PStart.Y) / float64(in.Points)

	for x := int32(0); x < in.BlockSize; x++ {
		for y := int32(0); y < in.BlockSize; y++ {
			cReal := in.PStart.X + float64(x+in.BlockSize*in.XBlock)*xStep
			cImag := in.PStart.Y + float64(y+in.BlockSize*in.YBlock)*yStep
			c := complex(cReal, cImag)
			z := complex(0, 0)
			curIters := in.MaxIters
			for i := int32(1); i < in.MaxIters; i++ {
				z = cmplx.Pow(z, 2) + c
				if real(z)+imag(z) > 4 {
					curIters = i
					break
				}
			}
			br.Results = append(br.Results, curIters)
		}
	}

	return br, nil
}

func (s *server) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	hr := new(pb.HealthCheckResponse)
	hr.Status = pb.HealthCheckResponse_SERVING
	return hr, nil
}

func main() {
	log.Printf("Starting mandel worker: version=%s build=%s", Version, Build)

	lis, err := net.Listen("tcp", ":28000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMandelServiceServer(s, &server{})
	pb.RegisterHealthServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
