Build it
----------

```
go get -d github.com/hasiotis/mandelbrot/v3
cd $GOPATH/src/github.com/hasiotis/mandelbrot/v3
make
```

Run it
----------

On a terminal run:

```
./mandelbrot
```

on a second terminal run:
```
curl -s http://localhost:8080 -o mandelbrot.png
eog mandelbrot.png
```
