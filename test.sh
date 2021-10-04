#!/bin/bash

go test ./pkg/... -coverprofile cover.out.tmp
cat cover.out.tmp | grep -v "_mock.go" > cover.out
go tool cover -html=cover.out
