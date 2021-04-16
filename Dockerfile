# syntax=docker/dockerfile:1.2

FROM golang:1.16-alpine as build

RUN apk add --no-cache gcc musl-dev

WORKDIR $GOPATH/github.com/bitmark-inc/autonomy-pod-controller

ADD . .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go install github.com/bitmark-inc/autonomy-pod-controller

# ---

FROM alpine:3.10.3
ARG dist=0.0
COPY --from=build /go/bin/autonomy-pod-controller /autonomy-pod-controller
