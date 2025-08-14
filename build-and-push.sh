#!/usr/bin/env bash
set -euo pipefail

COMPONENTS=("api" "worker" "scheduler")
API_IMAGE="ypeskov/orgfin-api-go"
WORKER_IMAGE="ypeskov/orgfing-worker-go"
SCHEDULER_IMAGE="ypeskov/orgfin-scheduler-go"

API_DOCKERFILE="Dockerfile"
WORKER_DOCKERFILE="Dockerfile.worker"
SCHEDULER_DOCKERFILE="Dockerfile.scheduler"

get_image_name() {
    local component=$1
    case $component in
        "api") echo "$API_IMAGE" ;;
        "worker") echo "$WORKER_IMAGE" ;;
        "scheduler") echo "$SCHEDULER_IMAGE" ;;
    esac
}

get_dockerfile() {
    local component=$1
    case $component in
        "api") echo "$API_DOCKERFILE" ;;
        "worker") echo "$WORKER_DOCKERFILE" ;;
        "scheduler") echo "$SCHEDULER_DOCKERFILE" ;;
    esac
}

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
    local component=$1
    local tag=$2
    local platform_option="--platform=linux/arm64"

    local image_name=$(get_image_name $component)
    local dockerfile=$(get_dockerfile $component)

    local build_command="docker build --target prod "
    
    build_command+="--no-cache $platform_option"
    build_command+=" -f $dockerfile -t $image_name:$tag ."
    echo "Building $component: $build_command"

    eval "$build_command"
}

build_all() {
    local tag=$1
    
    # Clear version.txt file
    > version.txt
    
    for component in "${COMPONENTS[@]}"; do
        echo "Building $component..."
        build_and_tag $component $tag
        echo "$component: $tag" >> version.txt
    done
}

push_all() {
    local tag=$1
    
    for component in "${COMPONENTS[@]}"; do
        local image_name=$(get_image_name $component)
        echo "Pushing $component..."
        docker push $image_name:$tag
    done
}


if [[ $# -eq 0 ]]; then
    tag=$(get_tag)
    build_all $tag
elif [[ $1 == "push" ]]; then
    tag=$(get_tag "${@:2}")
    build_all $tag
    push_all $tag
elif [[ $1 == "help" || $1 == "--help" || $1 == "-h" ]]; then
    echo "Usage:"
    echo "$0                                           # Build and tag all components (api/worker/scheduler) :latest"
    echo "$0 <tag>                                     # Build and tag all components as <tag>"
    echo "$0 push                                      # Build, tag and push all components :latest to Docker Hub"
    echo "$0 push <tag>                                # Build, tag as <tag> and push all components to Docker Hub"
    echo "$0 push [--platform=<platform>]              # Build for optional platform, tag as :latest, and push all components to Docker Hub"
    echo "$0 push <tag> [--platform=<platform>]        # Build for optional platform, tag as <tag>, and push all components to Docker Hub"
else
    # Build with specified tag
    tag=$(get_tag "$@")
    build_all $tag
fi
