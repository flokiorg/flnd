set shell := ["bash", "-c"]

IMAGE := "flokiorg/flnd:local"

# List available recipes
default:
    @just --list

# Build a runnable flnd docker image from the local working tree (not a git
# checkout), so lokihub/flspd dev environments can run against in-progress
# flnd changes. See dev.Dockerfile. Pulls in ../walletd and
# ../flokicoin-neutrino as Go workspace members too, since they evolve in
# lockstep with flnd on this fork.
docker-build:
    docker build -f dev.Dockerfile \
        --build-context walletd=../walletd \
        --build-context flokicoin-neutrino=../flokicoin-neutrino \
        -t {{IMAGE}} .
