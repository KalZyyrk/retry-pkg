#!/bin/bash
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
MIN=80.0

if (( $(echo "$COVERAGE < $MIN" | bc -l) )); then
    echo "Coverage too low: $COVERAGE% (min: $MIN%)"
    exit 1
fi