REVISION = $(shell git describe --tags)
$(info    Make orax-cli $(REVISION))

LDFLAGS := "-s -w -X gitlab.com/oraxpool/orax-cli/common.Version=$(REVISION)

# Set prod endpoints
LDFLAGS_PROD := $(LDFLAGS) -X gitlab.com/oraxpool/orax-cli/api.oraxAPIBaseURL=https://api.oraxpool.com
LDFLAGS_PROD := $(LDFLAGS_PROD) -X gitlab.com/oraxpool/orax-cli/ws.orchestratorURL=wss://orchestrator.oraxpool.com/miner
LDFLAGS_PROD := $(LDFLAGS_PROD)"

# Set staging endpoints
LDFLAGS_STAGING := $(LDFLAGS) -X gitlab.com/oraxpool/orax-cli/api.oraxAPIBaseURL=https://api.staging.oraxpool.com
LDFLAGS_STAGING := $(LDFLAGS_STAGING) -X gitlab.com/oraxpool/orax-cli/ws.orchestratorURL=wss://orchestrator.staging.oraxpool.com/miner
LDFLAGS_STAGING := $(LDFLAGS_STAGING)"

prod: orax-cli.app orax-cli.exe orax-cli orax-cli.arm64
staging: orax-cli-staging.app orax-cli-staging.exe orax-cli-staging orax-cli-staging.arm64

BUILD_FOLDER := "build"
PROD_BUILD_FOLDER := "$(BUILD_FOLDER)/prod"
STAGING_BUILD_FOLDER := "$(BUILD_FOLDER)/staging"

# Prod targets
orax-cli.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).app
	cp $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).app $(PROD_BUILD_FOLDER)/orax-cli.app
orax-cli.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).exe
	cp $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).exe $(PROD_BUILD_FOLDER)/orax-cli.exe
orax-cli:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION)
	cp $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION) $(PROD_BUILD_FOLDER)/orax-cli
orax-cli.arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).arm64
	cp $(PROD_BUILD_FOLDER)/orax-cli-$(REVISION).arm64 $(PROD_BUILD_FOLDER)/orax-cli.arm64

# Staging targets
orax-cli-staging.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.app
	cp $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.app $(STAGING_BUILD_FOLDER)/orax-cli-staging.app
orax-cli-staging.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.exe
	cp $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.exe $(STAGING_BUILD_FOLDER)/orax-cli-staging.exe
orax-cli-staging:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging
	cp $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging $(STAGING_BUILD_FOLDER)/orax-cli-staging
orax-cli-staging.arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.arm64
	cp $(STAGING_BUILD_FOLDER)/orax-cli-$(REVISION)-staging.arm64 $(STAGING_BUILD_FOLDER)/orax-cli-staging.arm64

.PHONY: clean

clean:
	rm -f orax-cli
	rm -rf build
