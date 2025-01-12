#!/bin/bash
set -e

OUTPUT_DIR="build"
mkdir -p $OUTPUT_DIR

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/gama-srv-64w.exe ./cmd/srv/

# Windows (32-bit)
GOOS=windows GOARCH=386 go build  -o $OUTPUT_DIR/gama-srv-32w.exe ./cmd/srv/

# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build  -o $OUTPUT_DIR/gama-srv ./cmd/srv/

# Linux (32-bit)
GOOS=linux GOARCH=386 go build -o $OUTPUT_DIR/gama-srv32 ./cmd/srv/

chmod +x build/*

echo "Builds completados en $OUTPUT_DIR"
