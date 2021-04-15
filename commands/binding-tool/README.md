# binding tool

This is a tool for testing autonomy pod controller by CLI.

## Pre-requisite

- go 1.16


## Build

```
go build
```

## Run

First, create a `config.yaml` by copy `config.yaml.sample` and configure it properly. Currently, the user needs to generate the `client_jwt` from autonomy API by themselves.

The usage of the tool could be displayed by:

```
./binding-tool -h
```


