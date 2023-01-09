package main

import (
	"fmt"

	"github.com/crewlinker/protoc-gen-appsync-go/internal/generator"
	"go.uber.org/zap"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gp *protogen.Plugin) error {
		gp.SupportedFeatures = 1 // seems to enable support for optional fields
		logs, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("failed to setup logging: %w", err)
		}

		logs = logs.Named("protoc-gen-appsync-go")

		gen := generator.New(logs)
		for _, name := range gp.Request.FileToGenerate {
			gf := gp.FilesByPath[name]
			if len(gf.Services) < 1 {
				continue // without services there is nothing to build a graphql schema for
			}

			logs.Info("found file with services", zap.Int("num_services", len(gf.Services)))
			gqlf := gp.NewGeneratedFile(fmt.Sprintf("%s.graphql", gf.GeneratedFilenamePrefix), gf.GoImportPath)
			if err = gen.GenerateSchema(gqlf); err != nil {
				return fmt.Errorf("failed to generate schema: %w", err)
			}
		}

		return nil
	})
}
