REVISION = $(shell git describe --tags)
$(info    Make orax-cli $(REVISION))

LDFLAGS := "-X gitlab.com/pbernier3/orax-cli/common.Version=$(REVISION)

# Set prod endpoints
LDFLAGS_PROD := $(LDFLAGS) -X gitlab.com/pbernier3/orax-cli/api.oraxAPIBaseURL=https://api.oraxpool.com
LDFLAGS_PROD := $(LDFLAGS_PROD) -X gitlab.com/pbernier3/orax-cli/ws.orchestratorURL=wss://orchestrator.oraxpool.com
LDFLAGS_PROD := $(LDFLAGS_PROD)"

# Set test endpoints
LDFLAGS_TEST := $(LDFLAGS) -X gitlab.com/pbernier3/orax-cli/api.oraxAPIBaseURL=https://orax-api.luciap.ca
LDFLAGS_TEST := $(LDFLAGS_TEST) -X gitlab.com/pbernier3/orax-cli/ws.orchestratorURL=wss://orchestrator.luciap.ca
LDFLAGS_TEST := $(LDFLAGS_TEST)"

dist: orax-cli.app orax-cli.exe orax-cli
dist-test: orax-cli-test.app orax-cli-test.exe orax-cli-test

# Prod targets
orax-cli.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o orax-cli-$(REVISION).app
orax-cli.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o orax-cli-$(REVISION).exe
orax-cli:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_PROD) -o orax-cli-$(REVISION)

# Test targets
orax-cli-test.app:
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o orax-cli-$(REVISION)-test.app
orax-cli-test.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o orax-cli-$(REVISION)-test.exe
orax-cli-test:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS_TEST) -o orax-cli-$(REVISION)-test

.PHONY: clean

clean:
	rm -f ./orax-cli ./orax-cli-v*
