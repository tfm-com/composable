package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/0xTFM/composable-cosmos/app"
	cmd "github.com/0xTFM/composable-cosmos/cmd/picad/cmd"
	cmdcfg "github.com/0xTFM/composable-cosmos/cmd/picad/config"
)

func main() {
	cmdcfg.SetupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
