package simplev1

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
	"Query.echo", "Query.echoV2", "Query.latestVersion", "Query.listProfiles",
}

// SimpleServiceResolver describes the resolver implementation using connect signatures.
type SimpleServiceResolver interface {
	Echo(context.Context, *connectgo.Request[EchoRequest]) (*connectgo.Response[EchoResponse], error)

	ListProfiles(context.Context, *connectgo.Request[ListProfilesRequest]) (*connectgo.Response[ListProfilesResponse], error)

	Version(context.Context, *connectgo.Request[VersionRequest]) (*connectgo.Response[VersionResponse], error)
}

// ResolveSimpleService resolves graphql calls
func ResolveSimpleService(ctx context.Context, h SimpleServiceResolver, typName, fldName string, args []byte) (data []byte, err error) {
	qualifier := fmt.Sprintf("%s.%s", typName, fldName)
	switch qualifier {

	case "Query.echo":
		var in EchoRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.Echo(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call handler: %w", err)
		}

		if data, err = protojson.Marshal(resp.Msg); err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}

		return data, nil

	case "Query.echoV2":
		var in EchoRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.Echo(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call handler: %w", err)
		}

		if data, err = protojson.Marshal(resp.Msg); err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}

		return data, nil

	case "Query.latestVersion":
		var in VersionRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.Version(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to call handler: %w", err)
		}

		if data, err = protojson.Marshal(resp.Msg); err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}

		return data, nil

	case "Query.listProfiles":
		var in ListProfilesRequest
		if err := protojson.Unmarshal(args, &in); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		req := connectgo.NewRequest(&in)

		resp, err := h.ListProfiles(ctx, req)
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
