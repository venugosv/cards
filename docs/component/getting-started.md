# Getting Started

## Prerequisites
Required
 -  [Go 1.14](https://golang.org/doc/install)
 -  [golangci-lint 1.20.1](https://github.com/golangci/golangci-lint#install)
 -  [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway#installation)
 -  proto3 and grpc (brew install protoc-gen-go grpc)

Optional
 -  [Goland IDE](https://www.jetbrains.com/go/download/)
 -  grpcurl (brew install grpcurl) helpful gRPC curl alternative
 -  make (brew install make)
 -  [prototool](https://github.com/uber/prototool/blob/dev/docs/install.md)

*Note*: If you followed the instructions on [Onboarding and Systems](https://github.com/anzx/fabric-docs/blob/master/website/docs/general/onboarding/index.md), you should already have everything required.

## Building & running the application
There are a number of ways to build the application

- You can execute the typical go build command `go build ./...`
- You can execute the Makefile step `make build`

There are similar ways to run the application
- You can execute the typical go run command `go run ./...`, however the application expects program arguments...
- You can execute the Makefile step `make run-cards` or `make run-cardcontrols` to start cards or cardcontrols service

The typical arguments for execution are in config files `-config deployment/local.yaml` You will need an environment variable for `FABRIC_CLIENT_ID` chat with a fabric team member to get a secret.

For other commands, please refer to [Makefile](../../Makefile)

## Where to get started with development
For an overview of the repository structure, [read the repository map](repository-map.md).
