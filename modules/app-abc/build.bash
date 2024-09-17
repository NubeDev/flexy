#!/bin/bash

# Check if correct number of arguments are passed
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <build_name> <zip_name> <destination_path>"
    exit 1
fi

# Assign arguments to variables
BUILD_NAME=$1
ZIP_NAME=$2
DEST_PATH=$3

# Set Go binary output name
GO_BINARY="${BUILD_NAME}"

# Build the Go binary
echo "Building Go project..."
go build -o "${GO_BINARY}"
if [ $? -ne 0 ]; then
    echo "Error: Go build failed."
    exit 1
fi
echo "Go project built successfully: ${GO_BINARY}"

# Find all .yaml files
echo "Finding .yaml files..."
YAML_FILES=$(find . -name "*.yaml")

# Create a zip with the Go binary and .yaml files
echo "Creating zip file: ${ZIP_NAME}.zip"
zip "${ZIP_NAME}.zip" "${GO_BINARY}" ${YAML_FILES}
if [ $? -ne 0 ]; then
    echo "Error: Failed to create zip file."
    exit 1
fi
echo "Zip file created: ${ZIP_NAME}.zip"

# Move the zip file to the destination path
echo "Moving zip file to ${DEST_PATH}"
#sudo mv "${ZIP_NAME}.zip" "${DEST_PATH}"
if [ $? -ne 0 ]; then
    echo "Error: Failed to move zip file to ${DEST_PATH}."
    exit 1
fi
echo "Zip file moved to: ${DEST_PATH}"

echo "Build and zip process completed successfully."
