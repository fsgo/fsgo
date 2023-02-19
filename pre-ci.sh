#!/bin/bash
cd $(dirname $0)
go mod tidy

pushd example/otel
go mod tidy
popd