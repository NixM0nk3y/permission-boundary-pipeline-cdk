package main

import (
	"fmt"
	"log"
	"permission-boundary-pipeline-cdk/pkg/stacks"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk"

	"github.com/aws/jsii-runtime-go"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	log.Print("Starting Pipeline Build")

	app := awscdk.NewApp(&awscdk.AppProps{
		AnalyticsReporting: jsii.Bool(false),
	})

	stackProps := stacks.PipelineStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	}

	err := envconfig.Process("cdk", &stackProps)

	if err != nil {
		log.Fatal(err.Error())
	}

	id := fmt.Sprintf("%s%sPipelineStack", strings.Title(stackProps.Tenant), strings.Title(stackProps.Environment))

	stacks.PipelineStack(app, id, &stackProps)

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil
}
