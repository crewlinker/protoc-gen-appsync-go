// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: examples/simple/v1/simple.proto

package simplev1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/crewlinker/cqrs/proto/examples/simple/v1"
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
	// SimpleServiceName is the fully-qualified name of the SimpleService service.
	SimpleServiceName = "examples.simple.v1.SimpleService"
)

// SimpleServiceClient is a client for the examples.simple.v1.SimpleService service.
type SimpleServiceClient interface {
	// Echo method returns a string argument
	Echo(context.Context, *connect_go.Request[v1.EchoRequest]) (*connect_go.Response[v1.EchoResponse], error)
}

// NewSimpleServiceClient constructs a client for the examples.simple.v1.SimpleService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSimpleServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) SimpleServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &simpleServiceClient{
		echo: connect_go.NewClient[v1.EchoRequest, v1.EchoResponse](
			httpClient,
			baseURL+"/examples.simple.v1.SimpleService/Echo",
			opts...,
		),
	}
}

// simpleServiceClient implements SimpleServiceClient.
type simpleServiceClient struct {
	echo *connect_go.Client[v1.EchoRequest, v1.EchoResponse]
}

// Echo calls examples.simple.v1.SimpleService.Echo.
func (c *simpleServiceClient) Echo(ctx context.Context, req *connect_go.Request[v1.EchoRequest]) (*connect_go.Response[v1.EchoResponse], error) {
	return c.echo.CallUnary(ctx, req)
}

// SimpleServiceHandler is an implementation of the examples.simple.v1.SimpleService service.
type SimpleServiceHandler interface {
	// Echo method returns a string argument
	Echo(context.Context, *connect_go.Request[v1.EchoRequest]) (*connect_go.Response[v1.EchoResponse], error)
}

// NewSimpleServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSimpleServiceHandler(svc SimpleServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/examples.simple.v1.SimpleService/Echo", connect_go.NewUnaryHandler(
		"/examples.simple.v1.SimpleService/Echo",
		svc.Echo,
		opts...,
	))
	return "/examples.simple.v1.SimpleService/", mux
}

// UnimplementedSimpleServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSimpleServiceHandler struct{}

func (UnimplementedSimpleServiceHandler) Echo(context.Context, *connect_go.Request[v1.EchoRequest]) (*connect_go.Response[v1.EchoResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("examples.simple.v1.SimpleService.Echo is not implemented"))
}
