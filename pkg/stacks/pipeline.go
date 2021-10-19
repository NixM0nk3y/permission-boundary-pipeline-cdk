package stacks

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"permission-boundary-pipeline-cdk/pkg/util"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/awscodepipelineactions"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/cloudformationinclude"
	"github.com/aws/aws-cdk-go/awscdk/pipelines"

	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

// go version to use to compile the CDK in the pipeline
const GOVERSION = "1.17.2"

type PipelineStackProps struct {
	Tenant       string            `envconfig:"TENANT" default:"openenterprise"`
	Environment  string            `envconfig:"ENVIRONMENT" default:"staging"`
	Application  string            `envconfig:"APPLICATION" default:"superapp4000"`
	GithubOrg    string            `envconfig:"GITHUB_ORG" default:"NixM0nk3y"`
	GithubRepo   string            `envconfig:"GITHUB_REPO" default:"permission-boundary-pipeline-cdk"`
	GithubBranch string            `envconfig:"GITHUB_BRANCH" default:"main"`
	StackProps   awscdk.StackProps ``
}

var conditionRestrictToRegions = &map[string]interface{}{
	"StringEquals": &map[string]interface{}{
		"aws:RequestedRegion": jsii.Strings(
			"us-east-1", // Allow North Virginia for CloudFront.
			"eu-west-1", // Europe.
		),
	},
}

//
// Jump through some hoops to generate a bootstrap stack using cdk tooling
// so we don't need to maintain a copy of a evolving CDK bootstrap stack
// template
func GenerateBootStrapTemplate(Qualifier string) *os.File {

	tmpTemplate, err := ioutil.TempFile(os.TempDir(), "cdk-bootstrap-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	cmd := exec.Command(
		//"CDK_NEW_BOOTSTRAP=1",
		"cdk",
		"bootstrap",
		"--qualifier", Qualifier,
		fmt.Sprintf("aws://%s/%s", *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
		"--require-approval", "never",
		fmt.Sprintf("--toolkit-stack-name=%s-CDKToolkit", Qualifier),
		"--cloudformation-execution-policies=arn:aws:iam::aws:policy/AdministratorAccess",
		"--show-template",
	)

	cmd.Env = append(os.Environ(),
		"CDK_NEW_BOOTSTRAP=1",
	)

	template, err := cmd.Output()
	if err != nil {
		log.Fatal("Cannot generate template", err)
	}

	if _, err = tmpTemplate.Write(template); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	// Close the file
	if err := tmpTemplate.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpTemplate
}

func PipelineStack(scope constructs.Construct, id string, props *PipelineStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	CdkQualifier := util.CalculateQualifier(props.Tenant, props.Application)
	log.Printf("Generated Qualifier: %s\n", CdkQualifier)

	stack := awscdk.NewStack(scope, &id, &sprops)

	// generate our permissions boundary
	permissionsBoundary := addPermissionsBoundary(stack, props, CdkQualifier)
	awscdk.NewCfnOutput(stack, jsii.String("PermissionsBoundaryArn"), &awscdk.CfnOutputProps{
		Value: permissionsBoundary.ManagedPolicyArn(),
	})

	token := awscdk.NewCfnParameter(stack, jsii.String("GithubToken"), &awscdk.CfnParameterProps{
		Type:   jsii.String("String"),
		NoEcho: jsii.Bool(true),
	})

	sourceArtifact := awscodepipeline.NewArtifact(jsii.String("sourceArtifact"))

	githubAction := awscodepipelineactions.NewGitHubSourceAction(&awscodepipelineactions.GitHubSourceActionProps{
		ActionName: jsii.String("GithubSource"),
		Owner:      jsii.String(props.GithubOrg),
		Repo:       jsii.String(props.GithubRepo),
		Branch:     jsii.String(props.GithubBranch),
		OauthToken: awscdk.SecretValue_PlainText(awscdk.Token_AsString(token.Value(), &awscdk.EncodingOptions{})),
		Output:     sourceArtifact,
	})

	cloudAssemblyArtifact := awscodepipeline.NewArtifact(jsii.String("cloudAssemblyArtifact"))

	deployAction := pipelines.NewSimpleSynthAction(&pipelines.SimpleSynthActionProps{
		CloudAssemblyArtifact: cloudAssemblyArtifact,
		SourceArtifact:        sourceArtifact,
		InstallCommands:       jsii.Strings("npm install aws-cdk -g", "cd $HOME/.goenv && git pull --ff-only && cd -", "goenv install "+GOVERSION, "goenv local "+GOVERSION),
		SynthCommand:          jsii.String("make ci/deploy/application"),
		EnvironmentVariables: &map[string]*awscodebuild.BuildEnvironmentVariable{
			"TENANT": {
				Type:  awscodebuild.BuildEnvironmentVariableType_PLAINTEXT,
				Value: jsii.String(props.Tenant),
			},
			"ENVIRONMENT": {
				Type:  awscodebuild.BuildEnvironmentVariableType_PLAINTEXT,
				Value: jsii.String(props.Environment),
			},
		},
	})

	pipelines.NewCdkPipeline(stack, jsii.String("CdkPipeline"), &pipelines.CdkPipelineProps{
		CloudAssemblyArtifact: sourceArtifact,
		SelfMutating:          jsii.Bool(false),
		CrossAccountKeys:      jsii.Bool(false),
		SourceAction:          githubAction,
		SynthAction:           deployAction,
	})

	deployAction.Project().Role().AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings("sts:AssumeRole"),
		Resources: jsii.Strings(
			fmt.Sprintf("arn:aws:iam::%s:role/cdk-%s-deploy-role-%s-%s", *awscdk.Aws_ACCOUNT_ID(), CdkQualifier, *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
			fmt.Sprintf("arn:aws:iam::%s:role/cdk-%s-file-publishing-role-%s-%s", *awscdk.Aws_ACCOUNT_ID(), CdkQualifier, *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
			fmt.Sprintf("arn:aws:iam::%s:role/cdk-%s-image-publishing-role-%s-%s", *awscdk.Aws_ACCOUNT_ID(), CdkQualifier, *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
			fmt.Sprintf("arn:aws:iam::%s:role/cdk-%s-lookup-role-%s-%s", *awscdk.Aws_ACCOUNT_ID(), CdkQualifier, *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
		),
	}))

	deployAction.Project().Role().AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings("iam:PassRole"),
		Resources: jsii.Strings(
			fmt.Sprintf("arn:aws:iam::%s:role/cdk-%s-cfn-exec-role-%s-%s", *awscdk.Aws_ACCOUNT_ID(), CdkQualifier, *awscdk.Aws_ACCOUNT_ID(), *awscdk.Aws_REGION()),
		),
	}))

	// create our bootstrap CDK stack with our qualifier
	bootstrapTemplate := GenerateBootStrapTemplate(CdkQualifier)
	defer os.Remove(bootstrapTemplate.Name())
	template := cloudformationinclude.NewCfnInclude(stack, jsii.String("BootStrap"), &cloudformationinclude.CfnIncludeProps{
		TemplateFile: jsii.String(bootstrapTemplate.Name()),
		Parameters: &map[string]interface{}{
			"Qualifier": CdkQualifier,
		},
	})

	// attach a PermissionsBoundary to the CF Role
	cfnExecRole := template.GetResource(jsii.String("CloudFormationExecutionRole"))
	cfnExecRole.AddPropertyOverride(jsii.String("PermissionsBoundary"), jsii.String(*permissionsBoundary.ManagedPolicyArn()))

	// lock down the roles to our pipeline
	pipelinePolicy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:     jsii.String("AllowCodebuild"),
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("sts:AssumeRole"),
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewArnPrincipal(deployAction.Project().Role().RoleArn()),
		},
	})

	// remove the account level trust and replace just with the pipeline role
	cfnFilePublishingRole := template.GetResource(jsii.String("FilePublishingRole"))
	cfnFilePublishingRole.AddPropertyDeletionOverride(jsii.String("AssumeRolePolicyDocument.Statement.0"))
	cfnFilePublishingRole.AddPropertyOverride(jsii.String("AssumeRolePolicyDocument.Statement.2"), pipelinePolicy.ToStatementJson())

	cfnImagePublishingRole := template.GetResource(jsii.String("ImagePublishingRole"))
	cfnImagePublishingRole.AddPropertyDeletionOverride(jsii.String("AssumeRolePolicyDocument.Statement.0"))
	cfnImagePublishingRole.AddPropertyOverride(jsii.String("AssumeRolePolicyDocument.Statement.2"), pipelinePolicy.ToStatementJson())

	cfnDeploymentActionRole := template.GetResource(jsii.String("DeploymentActionRole"))
	cfnDeploymentActionRole.AddPropertyDeletionOverride(jsii.String("AssumeRolePolicyDocument.Statement.0"))
	cfnDeploymentActionRole.AddPropertyOverride(jsii.String("AssumeRolePolicyDocument.Statement.2"), pipelinePolicy.ToStatementJson())

	cfnLookupRole := template.GetResource(jsii.String("LookupRole"))
	cfnLookupRole.AddPropertyDeletionOverride(jsii.String("AssumeRolePolicyDocument.Statement.0"))
	cfnLookupRole.AddPropertyOverride(jsii.String("AssumeRolePolicyDocument.Statement.3"), pipelinePolicy.ToStatementJson())

	return stack
}

//
// Initially cribbed from https://adrianhesketh.com/2021/09/02/secure-your-aws-ci-cd-pipelines-with-a-permissions-boundary/
//
func addPermissionsBoundary(stack constructs.Construct, props *PipelineStackProps, Qualifier string) (pb awsiam.ManagedPolicy) {

	boundaryNameTemplate := awscdk.Fn_Sub(jsii.String(fmt.Sprintf("%s-permissions-boundary-${AWS::AccountId}", Qualifier)), nil)
	boundaryArnTemplate := awscdk.Fn_Sub(jsii.String(fmt.Sprintf("arn:aws:iam::${AWS::AccountId}:policy/%s-permissions-boundary-${AWS::AccountId}", Qualifier)), nil)

	resourceApplicationRoleWildcard := fmt.Sprintf("arn:aws:iam::%s:role/%s%s*", *awscdk.Aws_ACCOUNT_ID(), strings.Title(props.Tenant), strings.Title(props.Environment))

	// Create a permission boundary.
	pb = awsiam.NewManagedPolicy(stack, jsii.String("PermissionsBoundary"), &awsiam.ManagedPolicyProps{
		ManagedPolicyName: boundaryNameTemplate,
		Description:       jsii.String("Permission boundary to limit permissions of roles created by CI/CD user."),
	})

	// Allow reading IAM information, and simulating policies.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowIAMReadOnly"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"iam:Get*",
			"iam:List*",
			"iam:SimulatePrincipalPolicy",
		),
		Resources: jsii.Strings("*"),
	}))

	// Allow services that need a wildcard resource ID because the resource path is unknown in advance e.g. API Gateway.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowServerlessServices"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"apigateway:*",
			"dynamodb:*",
			"ec2:CreateNetworkInterface",
			"ec2:DeleteNetworkInterface",
			"ec2:Describe*",
			"kms:*",
			"lambda:*",
			"logs:*",
			"s3:*",
			"secretsmanager:*",
			"ssm:*",
			"xray:*",
		),
		Resources:  jsii.Strings("*"),
		Conditions: conditionRestrictToRegions,
	}))

	// Allow CloudFormation deployment.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowCloudFormationDeployment"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"cloudformation:CreateStack",
			"cloudformation:DescribeStackEvents",
			"cloudformation:DescribeStackResources",
			"cloudformation:DescribeStackResource",
			"cloudformation:DescribeStacks",
			"cloudformation:GetTemplate",
			"cloudformation:ListStackResources",
			"cloudformation:UpdateStack",
			"cloudformation:ValidateTemplate",
			"cloudformation:DeleteStack",
		),
		Resources:  jsii.Strings("*"),
		Conditions: conditionRestrictToRegions,
	}))

	// Allow validation of any stack.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowValidationOfAnyStack"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"cloudformation:ValidateTemplate",
		),
		Resources:  jsii.Strings("*"),
		Conditions: conditionRestrictToRegions,
	}))

	// Allow passing any roles that start with the application name to Lambda.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowPassRoleToLambda"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"iam:PassRole",
		),
		Resources: jsii.Strings(resourceApplicationRoleWildcard),
		Conditions: &map[string]interface{}{
			"StringEquals": &map[string]interface{}{
				"iam:PassedToService": jsii.String("lambda.amazonaws.com"),
			},
		},
	}))

	// Deny permissions boundary alteration.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("DenyPermissionsBoundaryAlteration"),
		Effect: awsiam.Effect_DENY,
		Actions: jsii.Strings(
			"iam:CreatePolicyVersion",
			"iam:DeletePolicy",
			"iam:DeletePolicyVersion",
			"iam:SetDefaultPolicyVersion",
		),
		Resources: &[]*string{boundaryArnTemplate},
	}))

	// Deny removal of permissions boundary from any role.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("DenyPermissionsBoundaryRemoval"),
		Effect: awsiam.Effect_DENY,
		Actions: jsii.Strings(
			"iam:DeleteRolePermissionsBoundary",
		),
		Resources: jsii.Strings("*"),
	}))

	// Allow permissions boundaries to be applied.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowUpsertRoleIfPermBoundaryIsBeingApplied"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"iam:CreateRole",
			"iam:UpdateRole",
			"iam:AttachRolePolicy",
			"iam:PutRolePolicy",
			"iam:PutRolePermissionsBoundary",
			"iam:UpdateRoleDescription",
			"iam:UpdateAssumeRolePolicy",
		),
		Resources: jsii.Strings("*"),
		Conditions: &map[string]interface{}{
			"StringEquals": &map[string]interface{}{
				"iam:PermissionsBoundary": boundaryArnTemplate,
			},
		},
	}))

	// Allow roles and policys to be tagged
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowTagging"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"iam:TagPolicy",
			"iam:UntagPolicy",
			"iam:TagRole",
			"iam:UntagRole",
		),
		Resources: jsii.Strings("*"),
	}))

	// Allow roles to be deleted.
	pb.AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowDeleteRole"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"iam:DetachRolePolicy",
			"iam:DeleteRolePolicy",
			"iam:DeleteRole",
		),
		Resources: jsii.Strings("*"),
	}))

	return pb
}
