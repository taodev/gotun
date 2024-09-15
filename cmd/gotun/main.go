package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var (
	globalCtx  context.Context
	configPath string
)

var mainCommand = &cobra.Command{
	Use:              "gotun",
	PersistentPreRun: preRun,
}

func init() {
	mainCommand.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "set configuration file path")
}

func main() {
	if err := mainCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

func preRun(cmd *cobra.Command, args []string) {
	globalCtx = context.Background()
}
