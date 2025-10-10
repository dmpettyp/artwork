#!/bin/bash

# Script to create a new ImageGraph via the HTTP API

# Get name from command line argument, or use default
NAME="${1:-Test ImageGraph}"

echo "Creating ImageGraph with name: $NAME"
echo ""

curl -v -X POST http://localhost:8080/api/imagegraphs \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"$NAME\"}"

echo ""
