#!/bin/bash

# List all ImageGraphs
curl -s http://localhost:8080/imagegraphs | if command -v jq &> /dev/null; then jq; else cat; fi
