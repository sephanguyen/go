#!/bin/sh
#this script mean to be run inside the protoc_builder container


for module in 'bob' 'tom'; do
	echo "Generating Dart proto for ${module}";
	mkdir -p ${GENPROTO_OUT}/${module}/dart
	protoc -I=/protobuf/${module} -I=/usr/local/include --dart_out=grpc:${GENPROTO_OUT}/${module}/dart/ /usr/local/include/google/protobuf/*.proto /usr/local/include/googleapis/google/rpc/error_details.proto /protobuf/${module}/*.proto;
done
