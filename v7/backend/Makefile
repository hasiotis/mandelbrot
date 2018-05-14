include ../Makefile.defines

# Binary output file
BINARY=mandelbrot-backend

# Setup ldflags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

${BINARY}: main.go
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}

docker: ${BINARY}
	docker build -t ${BINARY}:${VERSION} .

deploy: docker
	kubectl apply -f k8s/
	kubectl set image deployment mandelbrot-backend backend=mandelbrot-backend:${VERSION}

install:
	go install ${LDFLAGS} -o ${BINARY}

clean:
	if [ -f ${BINARY} ]; then rm -rf ${BINARY}; fi

.PHONY: clean install
