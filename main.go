package main

import (
	"math/rand"
	"time"

	"github.com/loft-sh/devpod-provider-digitalocean/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
