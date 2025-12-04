.PHONY: build clean

_ := $(shell mkdir -p bin)
_ := $(shell go build -o bin/helpmakego github.com/iwahbe/helpmakego)

build: bin/pulumi-resource-file

bin/pulumi-resource-file: $(shell bin/helpmakego .)
	go build -o $@

clean:
	rm -rf bin
