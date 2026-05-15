package main

import (
	"fmt"

	"github.com/hanifanggawi/vessel/cmd"
	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	err := config.Init()
	if err != nil {
		fmt.Println(err.Error())
	}
	cmd.Execute()
}
