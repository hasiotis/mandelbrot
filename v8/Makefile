all:
	@echo "Building mandelbrot"
	@go get google.golang.org/grpc
	@(cd rpc;      make; cd ../)
	@(cd frontend; make; cd ../)
	@(cd backend;  make; cd ../)

docker:
	@echo "Building mandelbrot docker files"
	@(cd frontend; make docker; cd ../)
	@(cd backend;  make docker; cd ../)

deploy:
	@echo "Deploy on kubernetes (minikube)"
	@(cd frontend; make deploy; cd ../)
	@(cd backend;  make deploy; cd ../)

tools:
	@echo "Installing build tools"
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/golang/protobuf/protoc-gen-go

clean:
	@(cd frontend; make clean; cd ../)
	@(cd backend;  make clean; cd ../)
