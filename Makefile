version := "0.0.1"
git_hash := $(shell git rev-parse HEAD | cut -c1-8)

ifneq "${VERSION}" ""
	version := ${VERSION}
endif

default: build

build:
	go build -a -installsuffix cgo -ldflags "-s -w -X main.VERSION=${version} -s -w -X main.HASH=${git_hash}"

clean:
	go clean

test:
	go test

fmt:
	go fmt

build_docker:
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s -w -X main.VERSION=${version} -s -w -X main.HASH=${git_hash}"
	docker build -t spider-web .