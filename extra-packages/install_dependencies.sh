#!/bin/sh

if [ ! -f go.mod ]; then
    go mod init extra-packages
fi

while read line; do
    go get -v $line
done < packages.txt
