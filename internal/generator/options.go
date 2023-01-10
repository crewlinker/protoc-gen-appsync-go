package generator

import (
	appsyncv1 "github.com/crewlinker/protoc-gen-appsync-go/proto/appsync/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// FieldOptions returns our plugin specific options for a field. If the field has no options
// it returns nil.
func FieldOptions(f *protogen.Field) *appsyncv1.FieldOptions {
	opts, ok := f.Desc.Options().(*descriptorpb.FieldOptions)
	if !ok {
		return nil
	}
	ext, ok := proto.GetExtension(opts, appsyncv1.E_Field).(*appsyncv1.FieldOptions)
	if !ok {
		return nil
	}
	if ext == nil {
		return nil
	}
	return ext
}

// MethodOptions returns our plugin specific options for a method. If the field has no options
// it returns nil.
func MethodOptions(f *protogen.Method) *appsyncv1.MethodOptions {
	opts, ok := f.Desc.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return nil
	}
	ext, ok := proto.GetExtension(opts, appsyncv1.E_Method).(*appsyncv1.MethodOptions)
	if !ok {
		return nil
	}
	if ext == nil {
		return nil
	}
	return ext
}
