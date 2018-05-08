Build it
----------

```
go get -d github.com/hasiotis/mandelbrot/v4
cd $GOPATH/src/github.com/hasiotis/mandelbrot/v4
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

on a third terminal run:
```
curl -s http://localhost:8080 -o mandelbrot.png
eog mandelbrot.png
```
