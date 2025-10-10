#!/bin/bash

# List all ImageGraphs
curl -s http://localhost:8080/api/imagegraphs | if command -v jq &> /dev/null; then jq; else cat; fi
