#!/bin/bash

plugins=(
    "aggo"
    "protoc-gen-go-agkitex"
    "protoc-gen-go-aghertz"
    "protoc-gen-go-agserver"
    "protoc-gen-go-agservice"
    "protoc-gen-go-agapi"
    "gen-go-db"
)

for i in ${plugins[@]}
do
	echo $i
    cd $i
	go install
    cd -
done
