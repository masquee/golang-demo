#!/bin/bash

echo "=== Running K8s Demo with Detailed HTTP Logging ==="
echo "This will show you the actual HTTP requests and responses"
echo

go build -o k8s-logging-demo main-with-logging.go
./k8s-logging-demo