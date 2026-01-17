package main

import "github.com/R4yL-dev/glcmd/cmd/glcli/cmd"

var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
}
