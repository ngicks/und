#! /bin/bash
go test -coverprofile=./coverinfo ./... && go tool cover -html=coverinfo -o coverinfo.html