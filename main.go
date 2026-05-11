package main

import (
	"github.com/hanifanggawi/vessel/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
