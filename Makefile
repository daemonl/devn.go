ME="`id -u`:`id -g`"
VERSION=$(shell git describe --tags --candidates=1 --dirty)
DOCKER_TAG=daemonl/devn:${VERSION}
FLAGS=-X main.Version=$(VERSION)


bin/%: FORCE 
	mkdir -p bin
	go get -v github.com/daemonl/devn.go/cmd/$*
	go build -ldflags="$(FLAGS)" -o $@ github.com/daemonl/devn.go/cmd/$*


buildall: FORCE bin/devn-hooker bin/devn-run

docker-build: FORCE buildall
	echo "Docker Tag: ${DOCKER_TAG}"
	docker build -t ${DOCKER_TAG} .

docker-push: FORCE docker-build
	docker push ${DOCKER_TAG}

.PHONY: FORCE
FORCE:
