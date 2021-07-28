# syntax=docker/dockerfile:1.2

# SPDX-License-Identifier: ISC
# Copyright (c) 2019-2021 Bitmark Inc.
# Use of this source code is governed by an ISC
# license that can be found in the LICENSE file.

FROM golang:1.16-alpine as build

RUN apk add --no-cache gcc musl-dev

WORKDIR $GOPATH/github.com/bitmark-inc/autonomy-pod-controller

ADD . .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go install github.com/bitmark-inc/autonomy-pod-controller

# ---

FROM alpine:3.10.3
ARG dist=0.0

RUN apk add --no-cache curl

COPY --from=build /go/bin/autonomy-pod-controller /autonomy-pod-controller

ADD config.yaml.sample /.config/pod_controller.yaml
