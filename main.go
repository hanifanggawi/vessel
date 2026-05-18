package main

import (
	"fmt"

	"github.com/hanifanggawi/vessel/cmd"
	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/joho/godotenv"
)

// version is overridden at release time via -ldflags "-X main.version=...".
// In dev builds it stays "dev", which is also what gates loading a local
// .env: a released binary must never silently pick up a stray .env from its
// working directory and override the hosts/config paths.
var version = "dev"

func main() {
	if version == "dev" {
		godotenv.Load()
	}
	cmd.SetVersion(version)
	err := config.Init()
	if err != nil {
		fmt.Println(err.Error())
	}
	cmd.Execute()
}
