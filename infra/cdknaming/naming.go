package cdknaming

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

// Qualifier is used in our magefiles and infra file and duplicating it can lead to some really
// weird errors so we setup a dedicated package for it
var Qualifier = "ClProtoAS"

// InstanceName retrieves the instance name from the context or an empty string
func InstanceName(app awscdk.App) string {
	v, _ := app.Node().TryGetContext(jsii.String("instance")).(string)
	return v
}
