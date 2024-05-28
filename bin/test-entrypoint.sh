#!/bin/bash -xe

export PATH="$(pwd):$PATH"
echo "Path: $PATH"

echo "Running go tests"
echo "Current dir: $(pwd)"

mkdir -p output

go test --coverprofile=output/c.out -v ./... | tee output/junit.output

go-junit-report < output/junit.output > output/junit.xml

gocov convert output/c.out | gocov-xml > output/coverage.xml

rm output/junit.output
