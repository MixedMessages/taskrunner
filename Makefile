.PHONY: \
	all \
	deps \
	deps-test \
	build \
	install \
	test \
	lint \
	vet \


all: test

deps:
	go get -d -v github.com/MixedMessages/taskrunner/...

deps-test:
	go get -d -t -v github.com/MixedMessages/taskrunner/...

build: deps
	go build github.com/MixedMessages/taskrunner/...

install: deps
	go install github.com/MixedMessages/taskrunner/...

lint:
	go list ./... | grep -v /examples/ | xargs -L1 golint

test: deps-test
	go test $$(go list ./... | grep -v /examples/) -v -cover

vet:
	go vet -v $$(go list ./... | grep -v /examples/)
