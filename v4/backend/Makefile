# Binary output file
BINARY=mandelbrot-backend

# Build time info
VERSION=4.0.0
BUILD=`git rev-parse HEAD`

# Setup ldflags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.DEFAULT_GOAL: ${BINARY}

${BINARY}: main.go
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS} -o ${BINARY}

clean:
	if [ -f ${BINARY} ]; then rm -rf ${BINARY}; fi

.PHONY: clean install
