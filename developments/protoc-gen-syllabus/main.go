package main

import (
	//nolint
	"github.com/manabie-com/backend/developments/protoc-gen-syllabus/internal"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	opt := protogen.Options{}
	internal.Run(opt)
}
