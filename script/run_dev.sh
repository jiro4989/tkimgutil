#!/bin/bash

set -eu

find img/actor001/stands/right/ -type f |
	go run main.go version.go commands.go scale -s 100 |
	go run main.go version.go commands.go trim -x 40 -y 320 |
  sort |
	go run main.go version.go commands.go paste