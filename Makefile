# Meta tasks
# ----------

# Useful variables

# github oauth token
export TOKEN ?= AAAABBBB

# deployment environment
export ENVIRONMENT ?= production

# region
export AWS_REGION ?= eu-west-1

#
export AWS_ACCOUNT ?= 074705540277

export CODEBUILD_BUILD_NUMBER ?= 0
export CODEBUILD_RESOLVED_SOURCE_VERSION ?=$(shell git rev-list -1 HEAD --abbrev-commit)
export DATE=$(shell date -u '+%Y%m%d')

# Use a alternate CDK Qualifier to allow seperation of apps
export KMSID ?= AWS_MANAGED_KEY
export CDKQUALIFIER=$(shell jq -r .context.'"@aws-cdk/core:bootstrapQualifier"' < cdk.json)

# Output helpers
# --------------

TASK_DONE = echo "✓  $@ done"
TASK_BUILD = echo "🛠️  $@ done"

# ----------------
STACKS = $(shell find ./cmd/ -mindepth 1 -maxdepth 1 -type d)

.DEFAULT_GOAL := build

test:
	go test -v -p 1 ./...
	@$(TASK_BUILD)

bootstrap:
	CDK_NEW_BOOTSTRAP=1 cdk bootstrap --qualifier $(CDKQUALIFIER) aws://$(AWS_ACCOUNT)/$(AWS_REGION) --require-approval never --toolkit-stack-name=$(CDKQUALIFIER)-CDKToolkit --cloudformation-execution-policies=arn:aws:iam::aws:policy/AdministratorAccess --show-template
	@$(TASK_BUILD)

diff: diff/application
	@$(TASK_DONE)

synth: synth/application
	@$(TASK_DONE)

deploy: deploy/application
	@$(TASK_DONE)

synth/pipeline: build
	cdk synth --app ./pipeline --parameters GithubToken=$(TOKEN) --parameters FileAssetsBucketKmsKeyId=$(KMSID)
	@$(TASK_BUILD)

diff/pipeline: build
	cdk diff --app ./pipeline --parameters GithubToken=$(TOKEN) --parameters FileAssetsBucketKmsKeyId=$(KMSID)
	@$(TASK_BUILD)

deploy/pipeline: build
	cdk deploy --app ./pipeline --parameters GithubToken=$(TOKEN) --parameters FileAssetsBucketKmsKeyId=$(KMSID)
	@$(TASK_BUILD)

synth/application: build
	cdk synth --app ./application
	@$(TASK_BUILD)

diff/application: build
	cdk diff --app ./application
	@$(TASK_BUILD)

deploy/application: build
	cdk deploy --app ./application
	@$(TASK_BUILD)

ci/deploy/application: build
	cdk deploy --app ./application --ci true --require-approval never 
	@$(TASK_BUILD)

build: stacks/build
	@$(TASK_DONE)

.PHONY: stacks/build $(STACKS)

stacks/build: $(STACKS)
	@$(TASK_DONE)

$(STACKS):
	go build -v ./$@
	@$(TASK_BUILD)    
	
init: 
	go mod download
	@$(TASK_BUILD)

