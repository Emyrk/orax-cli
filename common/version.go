package common

// Version of the client. Set at build time (ldflag)
var Version string

func init() {
	if Version == "" {
		Version = "v0.1.0"
	}
}
