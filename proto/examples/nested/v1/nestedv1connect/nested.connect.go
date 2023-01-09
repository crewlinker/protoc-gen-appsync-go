// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: examples/nested/v1/nested.proto

package nestedv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/crewlinker/protoc-gen-appsync-go/proto/examples/nested/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// NestedServiceName is the fully-qualified name of the NestedService service.
	NestedServiceName = "examples.nested.v1.NestedService"
)

// NestedServiceClient is a client for the examples.nested.v1.NestedService service.
type NestedServiceClient interface {
	// KitchenSink method
	KitchenSink(context.Context, *connect_go.Request[v1.KitchenSinkRequest]) (*connect_go.Response[v1.KitchenSinkResponse], error)
}

// NewNestedServiceClient constructs a client for the examples.nested.v1.NestedService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewNestedServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) NestedServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &nestedServiceClient{
		kitchenSink: connect_go.NewClient[v1.KitchenSinkRequest, v1.KitchenSinkResponse](
			httpClient,
			baseURL+"/examples.nested.v1.NestedService/KitchenSink",
			opts...,
		),
	}
}

// nestedServiceClient implements NestedServiceClient.
type nestedServiceClient struct {
	kitchenSink *connect_go.Client[v1.KitchenSinkRequest, v1.KitchenSinkResponse]
}

// KitchenSink calls examples.nested.v1.NestedService.KitchenSink.
func (c *nestedServiceClient) KitchenSink(ctx context.Context, req *connect_go.Request[v1.KitchenSinkRequest]) (*connect_go.Response[v1.KitchenSinkResponse], error) {
	return c.kitchenSink.CallUnary(ctx, req)
}

// NestedServiceHandler is an implementation of the examples.nested.v1.NestedService service.
type NestedServiceHandler interface {
	// KitchenSink method
	KitchenSink(context.Context, *connect_go.Request[v1.KitchenSinkRequest]) (*connect_go.Response[v1.KitchenSinkResponse], error)
}

// NewNestedServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewNestedServiceHandler(svc NestedServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/examples.nested.v1.NestedService/KitchenSink", connect_go.NewUnaryHandler(
		"/examples.nested.v1.NestedService/KitchenSink",
		svc.KitchenSink,
		opts...,
	))
	return "/examples.nested.v1.NestedService/", mux
}

// UnimplementedNestedServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedNestedServiceHandler struct{}

func (UnimplementedNestedServiceHandler) KitchenSink(context.Context, *connect_go.Request[v1.KitchenSinkRequest]) (*connect_go.Response[v1.KitchenSinkResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("examples.nested.v1.NestedService.KitchenSink is not implemented"))
}
