package hosting

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/awslambdago"
	"github.com/aws/aws-cdk-go/awscdk/awslogs"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v3"
)

type HostingProps struct {
	Tenant           string                  ``
	Environment      string                  ``
	Appplication     string                  ``
	NestedStackProps awscdk.NestedStackProps ``
}

func HostingStack(scope constructs.Construct, id string, props *HostingProps) awscdk.Construct {

	construct := awscdk.NewConstruct(scope, &id)

	buildNumber, ok := os.LookupEnv("CODEBUILD_BUILD_NUMBER")
	if !ok {
		// default version
		buildNumber = "0"
	}

	sourceVersion, ok := os.LookupEnv("CODEBUILD_RESOLVED_SOURCE_VERSION")
	if !ok {
		sourceVersion = "unknown"
	}

	buildDate, ok := os.LookupEnv("BUILD_DATE")
	if !ok {
		t := time.Now()
		buildDate = t.Format("20060102")
	}

	// Go build options
	bundlingOptions := &awslambdago.BundlingOptions{
		GoBuildFlags: &[]*string{jsii.String(fmt.Sprintf(`-ldflags "-s -w
			-X api/pkg/version.Version=1.0.%s
			-X api/pkg/version.BuildHash=%s
			-X api/pkg/version.BuildDate=%s
			"`,
			buildNumber,
			sourceVersion,
			buildDate,
		)),
		},
		Environment: &map[string]*string{
			"GOARCH":      jsii.String("arm64"),
			"GO111MODULE": jsii.String("on"),
			"GOOS":        jsii.String("linux"),
		},
	}

	// webhook lambda
	apiLambda := awslambdago.NewGoFunction(construct, jsii.String("Lambda"), &awslambdago.GoFunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Entry:        jsii.String("resources/api/cmd/api"),
		Bundling:     bundlingOptions,
		Tracing:      awslambda.Tracing_ACTIVE,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Architectures: &[]awslambda.Architecture{
			awslambda.Architecture_ARM_64(),
		},
		Environment: &map[string]*string{
			"LOG_LEVEL": jsii.String("DEBUG"),
		},
		ModuleDir: jsii.String("resources/api/go.mod"),
	})

	apiLambda.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:     jsii.String("PermitParamGet"),
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("ssm:GetParametersByPath"),
		Resources: &[]*string{
			awscdk.Fn_Sub(jsii.String(fmt.Sprintf("arn:aws:ssm:eu-west-1:${AWS::AccountId}:parameter/%s/%s/%s/*", props.Tenant, props.Environment, props.Appplication)), nil),
		},
	}))

	//
	httpapi := awsapigatewayv2.NewHttpApi(construct, jsii.String("ApplicationAPI"), &awsapigatewayv2.HttpApiProps{})

	// POST
	apiIntegration := awsapigatewayv2integrations.NewLambdaProxyIntegration(&awsapigatewayv2integrations.LambdaProxyIntegrationProps{
		Handler:              apiLambda,
		PayloadFormatVersion: awsapigatewayv2.PayloadFormatVersion_VERSION_1_0(),
	})

	httpapi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Integration: apiIntegration,
		Path:        jsii.String("/version"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
	})

	httpapi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Integration: apiIntegration,
		Path:        jsii.String("/hello"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
	})

	return construct
}
