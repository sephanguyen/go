#!/bin/bash

set -eu



# Path to the YAML file
file_path="buf.gen.yaml"

# Get the value of the TRANSPORT_PATH environment variable
transport_path="${TRANSPORT_PATH:-}"

# Replace the TRANSPORT_PATH variable in the YAML content and save it to a temporary file
sed "s/\${TRANSPORT_PATH}/${transport_path}/g" "$file_path" > temp.yaml

# Replace the original YAML file with the temporary file
mv temp.yaml "$file_path"

# Get the value of the GRPC_BUFBUILD_OPTIONS environment variable
grpc_bufbuild_options="${GRPC_BUFBUILD_OPTIONS:-}"

# Replace the GRPC_BUFBUILD_OPTIONS variable in the YAML content and save it to a temporary file
sed "s/\${GRPC_BUFBUILD_OPTIONS}/${grpc_bufbuild_options}/g" "$file_path" > temp.yaml

# Replace the original YAML file with the temporary file
mv temp.yaml "$file_path"

echo "YAML file updated successfully!"


# Copy proto files to root
cp -a ./proto/* ./.

OUT_DIR="./pkg/manabuf-ts"

PROTO_DIR=${PROTO_DIR:-"./proto"}
TEAM=${TEAM:-""}

echo "PROTO_DIR: $PROTO_DIR"

mkdir -p ./pkg/manabuf-ts

PATH_ARGS=""

# Split string into array
set -- $PROTO_DIR

# and convert to string --path <dir> --path <dir> --path <dir>
for DIR in "$@"
do
    PATH_ARGS="$PATH_ARGS --path $DIR"
done

npx buf generate -v --debug $PATH_ARGS

chmod -R 777 "$OUT_DIR"

## Gen PlainMessageType for Response and Request

# Find all *_pb.ts files
files=$(find  -path '*_pb.ts')

# Iterate over each file
for file in $files; do
    # Search for classes ending in Respose or Request
    classes=$(grep -Eo 'class [A-Za-z0-9_]+(Response|Request)' "$file" | awk '{print $2}')

    # Iterate over each class
    for class in $classes; do
    
        if [[ -z "$class" ]]; then
            continue
        fi
        # Add lines for RequestType or ResponseType
        requestType="export type ${class}Type = PlainMessage<${class}>;"

        # Find response and request types is not exist before append to last line
        if ! grep -q "${class}Type" "$file"; then
            echo "" >> "$file"
            echo "$requestType" >> "$file"
        fi

    done
done


echo "Done."
