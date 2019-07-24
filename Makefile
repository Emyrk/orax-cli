REVISION = $(shell git describe --abbrev=0 --tags)
$(info    Make orax-cli $(REVISION))

orax-cli: 
	go build -ldflags "-X gitlab.com/pbernier3/orax-cli/common.Version=$(REVISION)" -o orax-cli

dist: orax-cli.app orax-cli.exe orax-cli-linux

orax-cli.app:
	env GOOS=darwin GOARCH=amd64 go build -o orax-cli-$(REVISION).app
orax-cli.exe:
	env GOOS=windows GOARCH=amd64 go build -o orax-cli-$(REVISION).exe
orax-cli-linux:
	env GOOS=linux GOARCH=amd64 go build -o orax-cli-$(REVISION)

.PHONY: clean

clean:
	rm -f ./orax-cli ./orax-cli-v*
