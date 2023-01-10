package simplev1

// ResolveSelectors list all type and field names that are resolved by the protobuf rpc methods. This
// is usefull to automate hooking up lambda functions to them in AppSync using a tool like AWS CDK.
var ResolveSelectors = []string{
	"Query.echo", "Query.echoV2", "Query.latestVersion", "Query.listProfiles",
}
