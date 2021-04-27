# Autonomy Pod Controller

![Run Testcases](https://github.com/bitmark-inc/autonomy-pod-controller/actions/workflows/run_testcases.yaml/badge.svg??branch=main)

This is controller of autonomy pod which enables a user to communicate with his
bitcoind via whisper protocol.

## Pre-requisite

- go 1.16


## Build

```
make pod-controller
```

## Run

First, create a `config.yaml` by copy `config.yaml.sample` and configure it properly. Then run:

```
make run-pod-controller
```

## Generate mock interfaces for testing

```
mockgen -source=store.go -destination store_mock.go -package main
```
