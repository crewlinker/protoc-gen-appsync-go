package nestedv1

import (
	"fmt"
	"context"
	"google.golang.org/protobuf/encoding/protojson"
	connectgo "github.com/bufbuild/connect-go"
)

//

// ResolveSelectors list all type and field names that are resolved by the protobuf rpc methods. This
// is usefull to automate hooking up lambda functions to them in AppSync using a tool like AWS CDK.
var ResolveSelectors = []string{
	"Post.related", "Query.posts",
}

// PostServiceResolver describes the resolver implementation using connect signatures.
type PostServiceResolver interface {
	Posts(context.Context, *connectgo.Request[PostsRequest]) (*connectgo.Response[PostsResponse], error)

	RelatedPosts(context.Context, *connectgo.Request[RelatedPostsRequest]) (*connectgo.Response[RelatedPostsResponse], error)
}

// ResolvePostService resolves graphql calls
func ResolvePostService(ctx context.Context, h PostServiceResolver, typName, fldName string, args []byte) (data []byte, err error) {
	qualifier := fmt.Sprintf("%s.%s", typName, fldName)
	switch qualifier {

	case "Post.related":
		var in RelatedPostsRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.RelatedPosts(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call handler: %w", err)
		}

		if data, err = protojson.Marshal(resp.Msg); err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}

		return data, nil

	case "Query.posts":
		var in PostsRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.Posts(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call handler: %w", err)
		}

		if data, err = protojson.Marshal(resp.Msg); err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}

		return data, nil
	default:
		return nil, fmt.Errorf("unsupported: %s", qualifier)
	}
}
