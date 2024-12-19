#!/bin/bash

if ! command -v docker &> /dev/null
then
  echo "Docker could not be found. Please install Docker to proceed."
  exit 1
fi

if [ -n "$MANIFEST_GITHUB_TOKEN" ]; then
  echo "MANIFEST_GITHUB_TOKEN is already set."
else
  read -p "Enter your MANIFEST_GITHUB_TOKEN: " MANIFEST_GITHUB_TOKEN
  if [ -z "$MANIFEST_GITHUB_TOKEN" ]; then
    echo "MANIFEST_GITHUB_TOKEN is required."
    exit 1
  fi
fi

current_dir=$(pwd)
docker run -v "$current_dir":/app -w /app -e MANIFEST_GITHUB_TOKEN=$MANIFEST_GITHUB_TOKEN ghcr.io/jonzudell/manifest/manifest:v0.0.4