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
	_ "net/http/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	pb "github.com/hasiotis/mandelbrot/v7/rpc"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type blockResult struct {
	blockX    int
	blockY    int
	Rectangle [blockSize][blockSize]uint8
}

const (
	pStart    complex128 = (-2.0 - 1.5i)
	pEnd      complex128 = (+0.6 + 1.5i)
	blockSize int        = 32
)

type config struct {
	Points        int
	MaxIters      int
	RedisServer   string
	BackendServer string
}

var (
	Version string
	Build   string
	mux     sync.Mutex
	C       config
	p       *pool.Pool
	b       *grpc.ClientConn
	c       pb.MandelServiceClient
	h       pb.HealthClient
	pOnline bool = false
	bOnline bool = false
)

func getCachedBlock(i int, j int) ([blockSize][blockSize]uint8, bool) {
	var cached bool = false
	var unserialized [blockSize][blockSize]uint8
	blockid := fmt.Sprintf("%03d%03d", i, j)
	if pOnline {
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
	if pOnline {
		serialized, err := json.Marshal(r)
		if err != nil {
			log.Printf("Serialize failed [%s]: blockid=%s\n", err, blockid)
		}
		mux.Lock()

		_, err = p.Cmd("HSET", "mandel", blockid, serialized).Int()
		if err != nil {
			log.Fatal(err)
		}
		mux.Unlock()
	}
}

func calculateMandel() *image.Gray {
	img := image.NewGray(image.Rect(0, 0, C.Points, C.Points))

	results := make(chan blockResult)
	var res blockResult

	for i := 0; i < int(C.Points/blockSize); i++ {
		for j := 0; j < int(C.Points/blockSize); j++ {
			go func(i int, j int) {
				var cached bool = false
				var ret blockResult
				ret.blockX = i
				ret.blockY = j
				if pOnline {
					ret.Rectangle, cached = getCachedBlock(i, j)
				}
				if !cached && bOnline {
					ps := &pb.ComplexPoint{real(pStart), imag(pStart)}
					pe := &pb.ComplexPoint{real(pEnd), imag(pEnd)}
					r, err := c.ComputeMandel(
						context.Background(),
						&pb.BlockRequest{ps, pe, int32(C.Points), int32(C.MaxIters), int32(blockSize), int32(i), int32(j)})
					if err != nil {
						log.Fatalf("Could not request compute: %v", err)
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

	for i := 0; i < int(C.Points/blockSize); i++ {
		for j := 0; j < int(C.Points/blockSize); j++ {
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
		log.Println("Unable to write image.")
	}
}

func redisConnect(retry bool) {
	var err error
	if !pOnline {
		p, err = pool.New("tcp", C.RedisServer, 10)
		if err == nil {
			pOnline = true
			log.Printf("Redis server is online")
		}
		return
	}

	if retry {
		p, err = pool.New("tcp", C.RedisServer, 10)
		if err != nil {
			pOnline = false
			log.Printf("Redis server is not reachable (retry): error=%s\n", err)
		} else {
			log.Printf("Redis server is online (retry)")
		}
	} else {
		mux.Lock()
		pong, err := p.Cmd("PING").Str()
		if err != nil || pong != "PONG" {
			pOnline = false
			log.Printf("Redis server is not reachable: error=%s\n", err)
		}
		mux.Unlock()
	}
}

func backendConnect(retry bool) {
	var err error
	if !bOnline {
		b, err = grpc.Dial(C.BackendServer, grpc.WithInsecure())
		if err == nil {
			c = pb.NewMandelServiceClient(b)
			h = pb.NewHealthClient(b)
			r, err := h.Check(context.Background(), &pb.HealthCheckRequest{"Check"})
			if err == nil && r.GetStatus().String() == "SERVING" {
				bOnline = true
				log.Printf("Backend server is online")
			}
		}
		return
	}

	if retry {
		log.Printf("  This is a retry")
		b, err = grpc.Dial(C.BackendServer, grpc.WithInsecure())
		if err != nil {
			bOnline = false
			log.Printf("Backend server is not reachable (retry): error=%s\n", err)
		} else {
			log.Printf("Backend server is online (retry)")
		}
	} else {
		h := pb.NewHealthClient(b)
		r, err := h.Check(context.Background(), &pb.HealthCheckRequest{"Check"})
		if err != nil || r.GetStatus().String() != "SERVING" {
			bOnline = false
			log.Printf("Backend server is not reachable: status=%s\n", r.GetStatus())
		}

	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	redisConnect(false)
	backendConnect(false)

	if pOnline || bOnline {
		img := calculateMandel()
		sendImage(w, img)
	} else {
		log.Printf("Both redis and backend servers are not available!\n")
	}
}

func readConfig() {
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("No configuration file loaded - using defaults")
	} else {
		log.Printf("Reading configuration from config file: configfile=config.yml")
	}

	err = viper.Unmarshal(&C)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	log.Printf("Configuration: Points=%d MaxIters=%d BackendServer=%s RedisServer=%s", C.Points, C.MaxIters, C.BackendServer, C.RedisServer)
}

func viewVersion(w http.ResponseWriter, r *http.Request) {
	versionOut := make(map[string]string)
	versionOut["version"] = Version
	versionOut["build"] = Build
	json.NewEncoder(w).Encode(versionOut)
}

func viewConfig(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(C)
}

func viewStatus(w http.ResponseWriter, r *http.Request) {
	statusOut := make(map[string]string)

	statusOut["redisConnection"] = strconv.FormatBool(pOnline)
	statusOut["backendConnection"] = strconv.FormatBool(bOnline)

	json.NewEncoder(w).Encode(statusOut)
}

func getConfig() {
	viper.SetConfigName("config")

	viper.AddConfigPath("/etc/mandelbrot-frontend/")
	viper.AddConfigPath("$HOME/.mandelbrot-frontend")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("MANDELBROT")

	viper.SetDefault("Points", 2048)
	viper.SetDefault("MaxIters", 256)
	viper.SetDefault("MaxIters", 256)

	viper.SetDefault("RedisServer", "localhost:6379")
	viper.SetDefault("BackendServer", "localhost:28000")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: filename=%s", e.Name)
		readConfig()
		redisConnect(true)
		backendConnect(true)
	})

	readConfig()
}

func main() {
	log.Printf("Starting mandelbrot frontend: version=%s build=%s\n", Version, Build)

	getConfig()

	http.HandleFunc("/", handler)
	http.HandleFunc("/version", viewVersion)
	http.HandleFunc("/config", viewConfig)
	http.HandleFunc("/status", viewStatus)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("Http server failed: msg=%s", err)
		}
	}()

	t := time.NewTicker(time.Second * 10)
	for {
		redisConnect(false)
		backendConnect(false)
		<-t.C
	}
}
