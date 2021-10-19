# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

VERSION=1.0

# Output helpers
# --------------

TASK_DONE = echo "✓  $@ done"
TASK_BUILD = echo "🛠️  $@ done"

export CODEBUILD_BUILD_NUMBER ?= 0
export CODEBUILD_RESOLVED_SOURCE_VERSION ?=$(shell git rev-list -1 HEAD --abbrev-commit)
export BUILD_DATE=$(shell date -u '+%Y%m%d')

all: test api/build

deps:
	go get -v  ./...

security/scan: build
	go run ./security/security.go
	@$(TASK_BUILD)
	
api/build:
	$(GOBUILD) -ldflags " \
		-X api/pkg/version.Version=${VERSION}.${CODEBUILD_BUILD_NUMBER} \
		-X api/pkg/version.BuildHash=${CODEBUILD_RESOLVED_SOURCE_VERSION} \
		-X api/pkg/version.BuildDate=${BUILD_DATE}" \
		-o ./bootstrap -v ./cmd/api
	@$(TASK_BUILD)

test: 
	@$(GOTEST) -v ./...
	@$(TASK_DONE)

clean:
	@$(GOCLEAN)
	@rm -f ./bootstrap
	@$(TASK_DONE)