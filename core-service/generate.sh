#!/bin/bash

set -e

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Navigate to project directory
cd "$(dirname "$0")"

# Create output directory
mkdir -p pkg/core

# Download googleapis if not exists
if [ ! -d "third_party/googleapis" ]; then
    echo "Downloading googleapis..."
    mkdir -p third_party
    cd third_party
    git clone --depth 1 https://github.com/googleapis/googleapis.git
    cd ..
fi

# Download protoc-gen-validate proto if not exists
if [ ! -d "third_party/validate" ]; then
    echo "Downloading protoc-gen-validate protos..."
    mkdir -p third_party/validate
    curl -L https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto -o third_party/validate/validate.proto
fi

echo "Generating protobuf code..."

# Generate protobuf files
protoc \
    -I. \
    -Ithird_party \
    -Ithird_party/googleapis \
    --go_out=pkg/core \
    --go_opt=paths=source_relative \
    --go-grpc_out=pkg/core \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=pkg/core \
    --grpc-gateway_opt=paths=source_relative \
    --grpc-gateway_opt=generate_unbound_methods=true \
    --validate_out="lang=go,paths=source_relative:pkg/core" \
    --openapiv2_out=pkg/core \
    --openapiv2_opt=allow_merge=true,merge_file_name=core \
    api/core/core.proto

echo "Protobuf generation completed successfully!"
