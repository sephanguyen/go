#!/bin/sh
#this script mean to be run inside the protoc_builder container
for module in /protobuf/*; do
	name=$(basename $module)
	echo "Generating Go proto for $name";
	OUT_DIR=${GENPROTO_OUT}/$name
	mkdir -p ${OUT_DIR}
	if [ $name = "yasuo" ] || [ $name = 'fatima' ] || [ $name = 'eureka' ]; then
		protoc -I=${module} -I=/protobuf/ -I=${GOPATH}/src \
			--gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,plugins=grpc,paths=source_relative:${OUT_DIR} ${module}/*.proto;
	else
		protoc -I=${module} -I=${GOPATH}/src \
			--gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,plugins=grpc,paths=source_relative:${OUT_DIR} ${module}/*.proto;
	fi;
done

