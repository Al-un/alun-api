#!/bin/sh

# Cheerful colouring
#   https://stackoverflow.com/a/61026468/4906586

# Must be root
if [ "$(id -u)" -ne 0 ]; then
    printf "\e[1;31m"
    echo "Please don't forget to sudo here"
    printf "\e[0m"
    exit 1
fi

# Constants
DOCKER_REPO=alunsng/alun

# Loop through Dockerfiles
for DOCKERFILE_NAME in *.Dockerfile
do

    # https://stackoverflow.com/a/428118/4906586
    APP_NAME=$(echo $DOCKERFILE_NAME | cut -d '.' -f 1)
    printf "\e[1;33m"
    echo "--- Build and pushing $APP_NAME ---"
    printf "\e[0m"

    docker build --file $DOCKERFILE_NAME --tag $DOCKER_REPO:$APP_NAME .
    docker push $DOCKER_REPO:$APP_NAME

done
