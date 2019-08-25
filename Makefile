REVISION = $(shell git describe --tags)
$(info    Make orax-cli $(REVISION))

LDFLAGS := "-X gitlab.com/pbernier3/orax-cli/common.Version=$(REVISION)

# Set prod endpoints
LDFLAGS_PROD := $(LDFLAGS) -X gitlab.com/pbernier3/orax-cli/api.oraxAPIBaseURL=https://api.oraxpool.com
LDFLAGS_PROD := $(LDFLAGS_PROD) -X gitlab.com/pbernier3/orax-cli/ws.orchestratorURL=wss://orchestrator.oraxpool.com/miner
LDFLAGS_PROD := $(LDFLAGS_PROD)"

# Set test endpoints
LDFLAGS_TEST := $(LDFLAGS) -X gitlab.com/pbernier3/orax-cli/api.oraxAPIBaseURL=https://orax-api.luciap.ca
LDFLAGS_TEST := $(LDFLAGS_TEST) -X gitlab.com/pbernier3/orax-cli/ws.orchestratorURL=wss://orchestrator.luciap.ca/miner
LDFLAGS_TEST := $(LDFLAGS_TEST)"

dist: orax-cli.app orax-cli.exe orax-cli
dist-test: orax-cli-test.app orax-cli-test.exe orax-cli-test

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

# Test targets
orax-cli-test.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-test.app
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-test.app $(BUILD_FOLDER)/orax-cli-test.app
orax-cli-test.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-test.exe
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-test.exe $(BUILD_FOLDER)/orax-cli-test.exe
orax-cli-test:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o $(BUILD_FOLDER)/orax-cli-$(REVISION)-test
	cp $(BUILD_FOLDER)/orax-cli-$(REVISION)-test $(BUILD_FOLDER)/orax-cli-test

.PHONY: clean

clean:
	rm -rf build
