#!/bin/bash
echo "=== Vizualisor Build (Linux) ==="
export CGO_ENABLED=1
echo "Building..."
go build -o vizualisor .
if [ $? -eq 0 ]; then
    echo "Build OK : vizualisor"
else
    echo "Build FAILED"
    exit 1
fi
