Build it
----------

```
go get -d github.com/hasiotis/mandelbrot/v5/...
cd $GOPATH/src/github.com/hasiotis/mandelbrot/v5
make tools
make
```

Run it
----------

On a terminal run:

```
./frontend/mandelbrot-frontend
```

On a second terminal run:

```
./backend/mandelbrot-backend
```

On a terminal terminal run:
```
docker run --name mandelbrot-redis --ulimit nofile=262144:262144 --network=host redis
```

And finally open another terminal and run:
```
curl -s http://localhost:8080 -o mandelbrot.png
eog mandelbrot.png
```

Congratulations you have just created a maintenance nightmare!
