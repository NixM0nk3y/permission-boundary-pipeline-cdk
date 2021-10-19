package main

import (
	"fmt"
	"log"
	"os"
	"permission-boundary-pipeline-cdk/pkg/stacks"
	"permission-boundary-pipeline-cdk/pkg/util"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"

	"github.com/kelseyhightower/envconfig"
)

func main() {

	log.Print("Starting Application Build")

	app := awscdk.NewApp(&awscdk.AppProps{
		AnalyticsReporting: jsii.Bool(false),
	})

	applicationProps := stacks.ApplicationProps{}

	err := envconfig.Process("cdk", &applicationProps)

	if err != nil {
		log.Fatal(err.Error())
	}

	applicationProps.Qualifier = util.CalculateQualifier(applicationProps.Tenant, applicationProps.Application)
	log.Printf("Generated Qualifier: %s\n", applicationProps.Qualifier)

	applicationProps.StackProps = awscdk.StackProps{
		Env: env(),
		Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
			Qualifier: jsii.String(applicationProps.Qualifier),
		}),
	}

	id := fmt.Sprintf("%s%sApplicationStack", strings.Title(applicationProps.Tenant), strings.Title(applicationProps.Environment))

	stacks.ApplicationStack(app, id, &applicationProps)

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
