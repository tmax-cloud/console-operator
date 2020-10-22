#!/usr/bin/env bash

set -e 
DOCKER_REGISTRY="tmaxcloudck"
PRODUCT="console-operator"
MAJOR_VERSION="4"
MINOR_VERSION="2"
PATCH_VERSION="0"
HOTFIX_VERSION="0"

#build docker image 
docker build -t ${DOCKER_REGISTRY}/${PRODUCT}:${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}.${HOTFIX_VERSION} -f ./Dockerfile .

docker push ${DOCKER_REGISTRY}/${PRODUCT}:${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}.${HOTFIX_VERSION}