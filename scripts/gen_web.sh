#!/bin/sh
#this script mean to be run inside the protoc_builder container

for module in 'bob' 'yasuo'; do
	echo "Generating Web proto for ${module}";
	OUT_DIR=${GENPROTO_OUT}/${module}/web
	mkdir -p ${OUT_DIR}

	if [ $module = "yasuo" ]; then
		protoc -I=/protobuf/${module} -I=/usr/local/include --proto_path=/protobuf/ --js_out=import_style=commonjs,binary:${OUT_DIR} \
		--grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:${OUT_DIR} /protobuf/bob/enum.proto /usr/local/include/google/protobuf/timestamp.proto /usr/local/include/google/protobuf/struct.proto /protobuf/${module}/*.proto;
	else 
		protoc -I=/protobuf/${module} -I=/usr/local/include --js_out=import_style=commonjs,binary:${OUT_DIR} \
		--grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:${OUT_DIR} /usr/local/include/google/protobuf/timestamp.proto /usr/local/include/google/protobuf/struct.proto /protobuf/${module}/*.proto;

	fi;
done
