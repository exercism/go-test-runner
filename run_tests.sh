#!/bin/bash

go test --json ./tests/1 > test.out
go run cmd/testoutput/main.go

go test --json ./tests/2 > test.out
go run cmd/testoutput/main.go
