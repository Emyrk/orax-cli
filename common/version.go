package common

// Version of the client. Set at build time (ldflag)
var Version string

func init() {
	if Version == "" {
		Version = "v1.0.0"
	}
}
