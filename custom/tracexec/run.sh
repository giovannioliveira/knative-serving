#!/bin/bash
go build ./tracexec.go
sudo nice -n -19 ./tracexec