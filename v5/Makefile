all:
	@echo "Building mandelbrot"
	@go get google.golang.org/grpc
	@(cd rpc;      make; cd ../)
	@(cd frontend; make; cd ../)
	@(cd backend;  make; cd ../)

tools:
	@echo "Installing build tools"
	@go get -u github.com/golang/dep/cmd/dep

clean:
	@(cd frontend; make clean; cd ../)
	@(cd backend;  make clean; cd ../)
