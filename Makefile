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

prod: orax-cli-darwin-amd64 orax-cli-windows-amd64.exe orax-cli-linux-amd64 orax-cli-linux-arm64 orax-cli-linux-arm7
staging: orax-cli-staging-darwin-amd64 orax-cli-staging-windows-amd64.exe orax-cli-staging-linux-amd64 orax-cli-staging-linux-arm64 orax-cli-staging-linux-arm7

BUILD_FOLDER := "build"
PROD_BUILD_FOLDER := "$(BUILD_FOLDER)/prod"
STAGING_BUILD_FOLDER := "$(BUILD_FOLDER)/staging"

# Prod targets
orax-cli-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-darwin-amd64-$(REVISION)
	cp $(PROD_BUILD_FOLDER)/orax-cli-darwin-amd64-$(REVISION) $(PROD_BUILD_FOLDER)/orax-cli-darwin-amd64
orax-cli-windows-amd64.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-windows-amd64-$(REVISION).exe
	cp $(PROD_BUILD_FOLDER)/orax-cli-windows-amd64-$(REVISION).exe $(PROD_BUILD_FOLDER)/orax-cli-windows-amd64.exe
orax-cli-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-linux-amd64-$(REVISION)
	cp $(PROD_BUILD_FOLDER)/orax-cli-linux-amd64-$(REVISION) $(PROD_BUILD_FOLDER)/orax-cli-linux-amd64
orax-cli-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-linux-arm64-$(REVISION)
	cp $(PROD_BUILD_FOLDER)/orax-cli-linux-arm64-$(REVISION) $(PROD_BUILD_FOLDER)/orax-cli-linux-arm64
orax-cli-linux-arm7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS_PROD) -o $(PROD_BUILD_FOLDER)/orax-cli-linux-arm7-$(REVISION)
	cp $(PROD_BUILD_FOLDER)/orax-cli-linux-arm7-$(REVISION) $(PROD_BUILD_FOLDER)/orax-cli-linux-arm7

# Staging targets
orax-cli-staging-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-staging-darwin-amd64-$(REVISION)
	cp $(STAGING_BUILD_FOLDER)/orax-cli-staging-darwin-amd64-$(REVISION) $(STAGING_BUILD_FOLDER)/orax-cli-staging-darwin-amd64
orax-cli-staging-windows-amd64.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-staging-windows-amd64-$(REVISION).exe
	cp $(STAGING_BUILD_FOLDER)/orax-cli-staging-windows-amd64-$(REVISION).exe $(STAGING_BUILD_FOLDER)/orax-cli-staging-windows-amd64.exe
orax-cli-staging-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-amd64-$(REVISION)
	cp $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-amd64-$(REVISION) $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-amd64
orax-cli-staging-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm64-$(REVISION)
	cp $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm64-$(REVISION) $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm64
orax-cli-staging-linux-arm7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags $(LDFLAGS_STAGING) -o $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm7-$(REVISION)
	cp $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm7-$(REVISION) $(STAGING_BUILD_FOLDER)/orax-cli-staging-linux-arm7


.PHONY: clean

clean:
	rm -f orax-cli
	rm -rf build
