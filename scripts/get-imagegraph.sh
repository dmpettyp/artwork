#!/bin/bash

# Script to fetch an ImageGraph by ID via the HTTP API

# Check if ID argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <imagegraph-id>"
  echo "Example: $0 550e8400-e29b-41d4-a716-446655440000"
  exit 1
fi

ID="$1"

echo "Fetching ImageGraph with ID: $ID"
echo ""

# Make the request and optionally pipe through jq for pretty printing
curl -v "http://localhost:8080/imagegraphs/$ID"

echo ""
