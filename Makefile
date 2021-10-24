GOCMD=go
GOBUILD=${GOCMD} build
GOTEST=${GOCMD} test
GOCLEAN=${GOCMD} clean
BINARY_NAME=b4t
VERSION=$(shell git rev-parse --short HEAD)

all: test build
build:
	${GOBUILD} -ldflags="-X 'main.version=${VERSION}'" -o ${BINARY_NAME} ./cmd/b4t/main.go
test:
	${GOTEST} ./...
dep:
	go mod vendor
	go mod tidy
clean:
	${GOCLEAN}
	rm ${BINARY_NAME}
run:
	${GOBUILD} -o ${BINARY_NAME} ./cmd/b4t/main.go
	./${BINARY_NAME}

