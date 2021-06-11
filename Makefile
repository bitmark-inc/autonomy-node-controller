# SPDX-License-Identifier: ISC
# Copyright (c) 2019-2021 Bitmark Inc.
# Use of this source code is governed by an ISC
# license that can be found in the LICENSE file.

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
	docker tag autonomy-wallet:pod-controller-$(dist) 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy-wallet:pod-controller-$(dist)

push:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy-wallet:pod-controller-$(dist)

test:
	go test ./...

clean:
	rm -rf bin
