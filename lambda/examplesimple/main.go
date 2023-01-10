package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bufbuild/connect-go"
	simplev1 "github.com/crewlinker/protoc-gen-appsync-go/proto/examples/simple/v1"
	"github.com/crewlinker/protoc-gen-appsync-go/proto/examples/simple/v1/simplev1connect"
)

type (
	// Input describes the input of a direct batch call from AWS AppSync that
	// we need for the generated resolve call to work
	Input = []struct {
		Arguments json.RawMessage `json:"arguments"`
		Source    json.RawMessage `json:"source"`
		Info      struct {
			FieldName           string `json:"fieldName"`
			ParentTypeName      string `json:"parentTypeName"`
			SelectionSetGraphQL string `json:"selectionSetGraphQL"`
		} `json:"info"`
	}

	// Output for a direct batch call from AWS AppSync
	Output = []map[string]any
)

// Handler handles lambda inputs
type Handler struct {
	impl simplev1.SimpleServiceResolver
}

// Handle direct lambda resolving from aws AppSync
func (h Handler) Handle(ctx context.Context, in Input) (out Output, err error) {
	log.Printf("Input: %+v", in)

	for _, item := range in {
		data, err := simplev1.ResolveSimpleService(ctx, h.impl, item.Info.ParentTypeName, item.Info.FieldName, item.Arguments)
		if err != nil {
			return nil, err
		}

		out = append(out, map[string]any{
			"data": json.RawMessage(data),
		})
	}

	log.Printf("Output: %+v", out)
	return
}

// Resolver implements the connect service
type Resolver struct {
	simplev1connect.UnimplementedSimpleServiceHandler
}

// Echo implements a method
func (r Resolver) Echo(ctx context.Context, req *connect.Request[simplev1.EchoRequest]) (resp *connect.Response[simplev1.EchoResponse], err error) {
	return connect.NewResponse(&simplev1.EchoResponse{Message: strings.ToUpper(req.Msg.Message)}), nil
}

// lambda entry point
func main() {
	lambda.Start((Handler{impl: Resolver{}}).Handle)
}
