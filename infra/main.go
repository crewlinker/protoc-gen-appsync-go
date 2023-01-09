package main

import (
	"os"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/protoc-gen-appsync-go/infra/appstack"
	"github.com/crewlinker/protoc-gen-appsync-go/infra/cdknaming"
)

func main() {
	defer jsii.Close()
	app, qualifier := awscdk.NewApp(nil), cdknaming.Qualifier
	instance := cdknaming.InstanceName(app)

	// deploy into the aws profile configured through --profile
	env, sp := &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}, &awscdk.DefaultStackSynthesizerProps{
		Qualifier: jsii.String(strings.ToLower(qualifier)),
	}

	// instance name scopes everything in the app stack
	stack := awscdk.NewStack(app, jsii.String(qualifier+"App"+instance), &awscdk.StackProps{
		Env:         env,
		Synthesizer: awscdk.NewDefaultStackSynthesizer(sp),
	})

	appstack.WithResources(stack)

	// synthesize the cloud assembly
	app.Synth(nil)
}
