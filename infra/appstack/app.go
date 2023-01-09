package appstack

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsappsync"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// WithResources builds the resources for the instanced app stack
func WithResources(s constructs.Construct) {
	for _, name := range []string{
		"Nested", "Simple",
	} {
		WithAppSync(s, name)
	}
}

// WithAppSync will setup an appsync api with a single lambda resolver
func WithAppSync(s constructs.Construct, name string) {
	s, lname := constructs.NewConstruct(s, jsii.String(name+"Graph")), strings.ToLower(name)

	// Setup the AppSync api
	api := awsappsync.NewCfnGraphQLApi(s, jsii.String("Api"), &awsappsync.CfnGraphQLApiProps{
		AuthenticationType: jsii.String("API_KEY"),
		Name:               jsii.String(*awscdk.Stack_Of(s).StackName() + "Graph"),
	})

	// setup an api key so we  can use the AWS query interface
	key := awsappsync.NewCfnApiKey(s, jsii.String("Key"), &awsappsync.CfnApiKeyProps{
		ApiId:       api.AttrApiId(),
		Description: jsii.String("Main API Key"),
		ApiKeyId:    jsii.String("MainApiKey"),
	})

	// Lambda resolver
	lambda := awslambda.NewFunction(s, jsii.String("Resolver"), &awslambda.FunctionProps{
		Code:         awslambda.AssetCode_FromAsset(jsii.String(filepath.Join("..", "lambda", "example"+lname, "pkg.zip")), nil),
		Handler:      jsii.String("bootstrap"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		LogRetention: awslogs.RetentionDays_ONE_DAY,
		Tracing:      awslambda.Tracing_ACTIVE,
		Timeout:      awscdk.Duration_Seconds(jsii.Number(50)),
	})

	role := awsiam.NewRole(s, jsii.String("ServiceRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("appsync.amazonaws.com"), nil),
	})

	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Resources: &[]*string{lambda.FunctionArn()},
		Actions:   jsii.Strings("lambda:InvokeFunction"),
	}))

	// Data source for the lambda resolver
	ds := awsappsync.NewCfnDataSource(s, jsii.String("LambdaSource"), &awsappsync.CfnDataSourceProps{
		ApiId:          api.AttrApiId(),
		Name:           jsii.String("LambdaSource"),
		Type:           jsii.String("AWS_LAMBDA"),
		ServiceRoleArn: role.RoleArn(),
		LambdaConfig: awsappsync.CfnDataSource_LambdaConfigProperty{
			LambdaFunctionArn: lambda.FunctionArn(),
		},
	})

	// read the schema file
	def, err := os.ReadFile(filepath.Join("..", "proto", "examples", lname, "v1", lname+".graphql"))
	if err != nil {
		panic("failed to load graphql definition: " + err.Error())
	}

	// define the schema for the api
	schema := awsappsync.NewCfnGraphQLSchema(s, jsii.String("Schema"), &awsappsync.CfnGraphQLSchemaProps{
		ApiId:      api.AttrApiId(),
		Definition: jsii.String(string(def)),
	})

	for _, typfield := range []string{
		// @TODO generate type/field listing for resolver
	} {
		typ, field, _ := strings.Cut(typfield, ".")
		awsappsync.NewCfnResolver(s, jsii.String(typ+field+"Resolver"), &awsappsync.CfnResolverProps{
			ApiId:          api.AttrApiId(),
			TypeName:       jsii.String(typ),
			FieldName:      jsii.String(field),
			DataSourceName: ds.AttrName(),
			MaxBatchSize:   jsii.Number(10), // enable batching for direct lambda
		}).AddDependency(schema)
	}

	// Output so we can test the AppSync api e2e
	awscdk.NewCfnOutput(s, jsii.String("HttpURL"), &awscdk.CfnOutputProps{
		Value: api.AttrGraphQlUrl(),
	})
	awscdk.NewCfnOutput(s, jsii.String("SecretKey"), &awscdk.CfnOutputProps{
		Value: key.AttrApiKey(),
	})
}
