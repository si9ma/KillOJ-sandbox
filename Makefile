all: build

build:
	go build -o kbox

clean:
	rm -rf kbox

install:
	go build -o $$GOPATH/bin/kbox
