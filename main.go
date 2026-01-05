package main

import (
	"github.com/zer0go/simple-load-balancer/cmd"
)

var Version = "development"

func main() {
	cmd.Execute(Version)
}
