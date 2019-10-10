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

BUILD_FOLDER="build"

# Prod targets
orax-cli.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(BUILD_FOLDER)/orax-cli-$(REVISION).app
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION).app $(BUILD_FOLDER)/orax-cli.app
orax-cli.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(BUILD_FOLDER)/orax-cli-$(REVISION).exe
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION).exe $(BUILD_FOLDER)/orax-cli.exe
orax-cli:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION) $(BUILD_FOLDER)/orax-cli
orax-cli.arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_PROD) -o $(BUILD_FOLDER)/orax-cli-$(REVISION).arm64
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION).arm64 $(BUILD_FOLDER)/orax-cli.arm64

# Staging targets
orax-cli-staging.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.app
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.app $(BUILD_FOLDER)/orax-cli-staging.app
orax-cli-staging.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.exe
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.exe $(BUILD_FOLDER)/orax-cli-staging.exe
orax-cli-staging:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_STAGING) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging $(BUILD_FOLDER)/orax-cli-staging
orax-cli-staging.arm64:
	env GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS_STAGING) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.arm64
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-staging.arm64 $(BUILD_FOLDER)/orax-cli-staging.arm64

.PHONY: clean

clean:
	rm -f orax-cli
	rm -rf build
