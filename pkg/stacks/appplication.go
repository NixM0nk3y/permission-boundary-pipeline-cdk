package stacks

import (
	"fmt"
	"permission-boundary-pipeline-cdk/pkg/hosting"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

type ApplicationProps struct {
	Tenant      string            `envconfig:"TENANT" default:"openenterprise"`
	Environment string            `envconfig:"ENVIRONMENT" default:"staging"`
	Application string            `envconfig:"APPLICATION" default:"superapp4000"`
	Qualifier   string            `envconfig:"QUALIFIER" default:"hnb659fds"`
	StackProps  awscdk.StackProps ``
}

func ApplicationStack(scope constructs.Construct, id string, props *ApplicationProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	hosting.HostingStack(stack, "Hosting", &hosting.HostingProps{
		Tenant:      props.Tenant,
		Environment: props.Environment,
	})

	// apply boundary to all roles within the stack
	boundary := awsiam.ManagedPolicy_FromManagedPolicyArn(
		stack,
		jsii.String("Boundary"),
		jsii.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s-permissions-boundary-%s", *awscdk.Aws_ACCOUNT_ID(), props.Qualifier, *awscdk.Aws_ACCOUNT_ID())),
	)
	awsiam.PermissionsBoundary_Of(stack).Apply(boundary)

	return stack
}
