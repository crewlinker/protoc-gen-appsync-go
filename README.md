# protoc-gen-appsync-go

Generate AWS AppSync-compatible GraphQL schema and resolver code from protobuf RPC definitions

## Goals

Allow an AppSync api to call a connec-go specced RPC implementations. It generates a graphql schema
and resolver implementation that can be uploaded to appsync to transparently call protobuf specced
backend code.

Ultimate goal is to support a subscripion that can transparently update a Relay connection on the
client side: https://relay.dev/docs/guided-tour/updating-data/graphql-subscriptions/ while the server
side is using resilient protobuf specced rpc.

## Design

- We use protojson to decode the appsync "arguments" into a protobuf "Request" type.
- Set http.Headers as passed from the appsync input
- The return message is protojson encoded as the return value (might not support top-level scalar returns)
- In case of nested resolvers. We decode the "source" field into a message and provide it through the context.Context
- Provide original request, including selectionset (parsed and unparsed) through the context.

## Why AppSync

- PRO: Build-in caching support (less valuable without nested resolvers)
- PRO: Build-in subscription support, easily update clients of changes wihout a custom websocket protocol
- PRO: Allows clients to use advanced graphql tooling (i.e Relay)

## Other approaches

Generating just resolvers from GraphQl schema.

- PRO: more expressive GraphQL schemas

Generating from protobuf RPC

- PRO: Easily add other ways to call the API (Rest)
- PRO: Protobuf has better tooling (buf, vs gqlgen)
- PRO: Comes with a de-facto validation project
- PRO: Better (proper) type support: 64-bit ints, -Infinity, Nan etc
- CON: No way to support shortcut map annotation `map<string,string>` only legacy map structures
- CON: Not clear if we can support nested resolvers (need to provide "parent" as a field, maybe annotate)

## Backlog

- [ ] SHOULD Test with nested resolvers, decode the AppSync "source" field into context.Context. Generate code to
      read it from the context.
- [ ] SHOULD test if it's feasible to validate the "source" (parent) context input to catch invalid calling
- [ ] SHOULD test calling a query with n+1 difficulty to check if batching works
- [ ] SHOULD test the use of AWS scalars for appsync: https://docs.aws.amazon.com/appsync/latest/devguide/scalars.html
