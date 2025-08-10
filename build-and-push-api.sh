#!/usr/bin/env bash
set -euo pipefail

TAG_PREFIX="ypeskov/api-go-orgfin"

get_tag() {
    local tag="latest"

    for arg in "$@"; do
        if [[ $arg != --platform=* ]] && [[ $arg != "push" ]]; then
            tag=$arg
            break
        fi
    done

    echo "$tag"
}

build_and_tag() {
    local tag=$1
    local platform_option="--platform=linux/arm64"

    local build_command="docker build --target prod "
    
    build_command+="--no-cache $platform_option"
    build_command+=" -t $TAG_PREFIX:$tag ."
    echo "$build_command"

    eval "$build_command"
    echo "$tag" > version.txt
}


if [[ $# -eq 0 ]]; then
    tag=$(get_tag)
    build_and_tag $tag
elif [[ $1 == "push" ]]; then
    tag=$(get_tag "${@:2}")
    build_and_tag $tag
    docker push $TAG_PREFIX:$tag

    echo "$tag" > ../version.txt
else
    echo "Usage:"
    echo "$0                                           # Build and tag :latest"
    echo "$0 push                                      # Build, tag and push :latest to Docker Hub"
    echo "$0 push <tag>                                # Build, tag as <tag> and push to Docker Hub"
    echo "$0 push [--platform=<platform>]              # Build for optional platform, tag as :latest, and push to Docker Hub"
    echo "$0 push <tag> [--platform=<platform>]        # Build for optional platform, tag as <tag>, and push to Docker Hub"
fi
