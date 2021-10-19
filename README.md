# Permissions Bound CDK deployment pipeline

Self Contained Demo of a CDK deployment pipeline incorporating a Permissions Boundary

# Stack Setup

```bash
TOKEN=$ghp_mytoken$  make deploy/pipeline
go build -v ./cmd/pipeline
permission-boundary-pipeline-cdk/cmd/pipeline
🛠️  cmd/pipeline done
go build -v ./cmd/application
permission-boundary-pipeline-cdk/cmd/application
🛠️  cmd/application done
✓  stacks/build done
✓  build done
cdk deploy --app ./pipeline --parameters GithubToken=$ghp_mytoken$ --parameters FileAssetsBucketKmsKeyId=AWS_MANAGED_KEY
2021/10/19 18:52:26 Starting Pipeline Build
2021/10/19 18:52:29 Generated Qualifier: 703ff7a19f

OpenenterpriseProductionPipelineStack: deploying...
[0%] start: Publishing 3ce87665dc3287fcec39a3f0d32d3073618d14628460ecd5ce0702c10d294cc2:current_account-current_region
[100%] success: Published 3ce87665dc3287fcec39a3f0d32d3073618d14628460ecd5ce0702c10d294cc2:current_account-current_region
OpenenterpriseProductionPipelineStack: creating CloudFormation changeset...

 ✅  OpenenterpriseProductionPipelineStack

Outputs:
OpenenterpriseProductionPipelineStack.BootstrapVersion = 8
OpenenterpriseProductionPipelineStack.BucketDomainName = cdk-703ff7a19f-assets-074705540277-eu-west-1.s3.eu-west-1.amazonaws.com
OpenenterpriseProductionPipelineStack.BucketName = cdk-703ff7a19f-assets-074705540277-eu-west-1
OpenenterpriseProductionPipelineStack.FileAssetKeyArn = AWS_MANAGED_KEY
OpenenterpriseProductionPipelineStack.ImageRepositoryName = cdk-703ff7a19f-container-assets-074705540277-eu-west-1
OpenenterpriseProductionPipelineStack.PermissionsBoundaryArn = arn:aws:iam::074705540277:policy/703ff7a19f-permissions-boundary-074705540277

Stack ARN:
arn:aws:cloudformation:eu-west-1:074705540277:stack/OpenenterpriseProductionPipelineStack/78ea0da0-2f40-11ec-8007-02213e755193
🛠️  deploy/pipeline done
```

The once the stack is deployed the pipeline is able to deploy the "application" via a codebuild pipeline.

The permissions boundary is configured to reflect the deployment requirements of a typical serverless application.

# License

MIT

## Useful commands

 * `make deploy`          deploy this stack to your default AWS account/region
