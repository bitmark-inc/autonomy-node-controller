.PHONY:

dist =

default: build

pod-controller:
	go build -o bin/pod-controller

run-pod-controller: pod-controller
	./bin/pod-controller -c config.yaml

bin: pod-controller

build-pod-controller:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy-wallet:pod-controller-$(dist) .

test:
	go test ./...

clean:
	rm -rf bin
