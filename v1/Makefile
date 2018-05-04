# Binary output file
BINARY=mandelbrot

# Build time info
VERSION=1.0.0
BUILD=`git rev-parse HEAD`

# Setup ldflags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.DEFAULT_GOAL: ${BINARY}

${BINARY}: main.go
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS} -o ${BINARY}

clean:
	if [ -f ${BINARY} ]; then rm -rf ${BINARY} *.png; fi

.PHONY: clean install
