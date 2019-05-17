package main

import "github.com/orientlu/lora-coordinator/cmd/lora-coordinator/cmd"

var version = "0.0.0" // set by compile

func main() {
	cmd.Execute(version)
}
