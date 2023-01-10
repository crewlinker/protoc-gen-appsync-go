package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bufbuild/connect-go"
	nestedv1 "github.com/crewlinker/protoc-gen-appsync-go/proto/examples/nested/v1"
	"github.com/crewlinker/protoc-gen-appsync-go/proto/examples/nested/v1/nestedv1connect"
	"github.com/samber/lo"
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
	resolver nestedv1.PostServiceResolver
}

// Handle direct lambda resolving from aws AppSync
func (h Handler) Handle(ctx context.Context, in Input) (out Output, err error) {
	log.Printf("Input: %+v", in)

	for _, item := range in {
		data, err := nestedv1.ResolvePostService(ctx, h.resolver, item.Info.ParentTypeName, item.Info.FieldName, item.Arguments)
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

// Resolver implements the post resolver
type Resolver struct {
	nestedv1connect.UnimplementedPostServiceHandler
	posts   map[string]*nestedv1.Post
	relates map[string][]string
}

func (r Resolver) RelatedPosts(
	ctx context.Context,
	req *connect.Request[nestedv1.RelatedPostsRequest],
) (resp *connect.Response[nestedv1.RelatedPostsResponse], err error) {
	// @TODO read the "parent" context

	return connect.NewResponse(&nestedv1.RelatedPostsResponse{}), nil
}

// posts returns posts
func (r Resolver) Posts(
	ctx context.Context,
	req *connect.Request[nestedv1.PostsRequest],
) (resp *connect.Response[nestedv1.PostsResponse], err error) {
	return connect.NewResponse(&nestedv1.PostsResponse{
		Posts: lo.Values(r.posts),
	}), nil
}

// lambda entry point
func main() {
	r := Resolver{
		posts: map[string]*nestedv1.Post{
			"post-1": {Id: "post-1"},
			"post-2": {Id: "post-2"},
			"post-3": {Id: "post-4"},
		},
		relates: map[string][]string{
			"post-2": {"post-1", "post-2"},
			"post-4": {"post-2"},
		},
	}

	lambda.Start((Handler{resolver: r}).Handle)
}
