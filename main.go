package main

import (
	"flag"
	"fmt"

	"github.com/crewlinker/protoc-gen-appsync-go/internal/generator"
	"go.uber.org/zap"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	queryMessage        = flag.String("query_message", "Query", "name of the message that describes the top-level Query type")
	mutationMessage     = flag.String("mutation_message", "Mutation", "name of the message that describes the top-level Mutation type")
	subscriptionMessage = flag.String("subscription_message", "Mutation", "name of the message that describes the top-level Mutation type")
)

func main() {
	flag.Parse()
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gp *protogen.Plugin) error {
		gp.SupportedFeatures = 1 // seems to enable support for optional fields
		logs, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("failed to setup logging: %w", err)
		}

		opts := &generator.Options{
			QueryMessageName:        *queryMessage,
			MutationMessageName:     *mutationMessage,
			SubscriptionMessageName: *subscriptionMessage,
		}

		gen, err := generator.New(logs, opts)
		if err != nil {
			return fmt.Errorf("failed to initialize generator: %w", err)
		}

		for _, name := range gp.Request.FileToGenerate {
			pf := gp.FilesByPath[name]
			if len(pf.Services) < 1 {
				continue // without services there is nothing to build a graphql schema for
			}

			logs.Info("found file with services", zap.Int("num_services", len(pf.Services)))
			resf, gqlf :=
				gp.NewGeneratedFile(fmt.Sprintf("%s.res.go", pf.GeneratedFilenamePrefix), pf.GoImportPath),
				gp.NewGeneratedFile(fmt.Sprintf("%s.graphql", pf.GeneratedFilenamePrefix), pf.GoImportPath)

			if err = gen.GenerateResolve(resf, pf); err != nil {
				return fmt.Errorf("failed to generate resolver code: %w", err)
			}

			if err = gen.GenerateSchema(gqlf, pf); err != nil {
				return fmt.Errorf("failed to generate schema: %w", err)
			}
		}

		return nil
	})
}
