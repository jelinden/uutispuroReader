#!/bin/bash
git pull
go build
killall uutispuroReader
export GOMAXPROCS=4 && nohup ./uutispuroReader -env=prod > uutispuroReader.log 2>&1&
