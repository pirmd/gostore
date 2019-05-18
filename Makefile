VERSION=`git describe --tags`
BUILD=`git rev-parse --short HEAD`

LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.build=${BUILD}"

build: deps
	go build ${LDFLAGS} 

check:
	go fmt ./...
	go vet ./...

test: deps-test check
	go test -cover -v ./...

test-cover: check
	go test -coverprofile=cover.out
	go tool cover -html=cover.out
	unlink cover.out

gostore.1:
	go generate ./...

install: deps gostore.1
	go install ${LDFLAGS}
	install -m 644 gostore.1 ${GOBIN}/../man/man1/

clean:
	go clean -i -x

deps:
	go get -d -v

deps-test:
	go test -i ./...

.PHONY: build check test test-cover deps deps-test install clean

# vim: set noexpandtab shiftwidth=8 softtabstop=0:
