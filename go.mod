module gitlab.com/oraxpool/orax-cli

go 1.12

require (
	github.com/FactomProject/btcutil v0.0.0-20160826074221-43986820ccd5 // indirect
	github.com/FactomProject/ed25519 v0.0.0-20150814230546-38002c4fe7b6 // indirect
	github.com/FactomProject/factom v0.0.0-20190712161128-5edf7247fc87
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.7.0
	github.com/google/flatbuffers v1.11.0
	github.com/gorilla/websocket v1.4.1
	github.com/goware/emailx v0.2.0
	github.com/manifoldco/promptui v0.3.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pegnet/LXRHash v0.0.0-20191028162532-138fe8d191a2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	github.com/stretchr/testify v1.4.0
	gitlab.com/oraxpool/orax-message v0.0.0-20190921191632-bfac1083c89e
	gopkg.in/resty.v1 v1.12.0
)

replace github.com/pegnet/LXRHash => /home/steven/go/src/github.com/pegnet/LXRHash
