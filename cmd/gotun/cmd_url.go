package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/cobra"
)

var commandTest = &cobra.Command{
	Use:   "url",
	Short: "test url",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(configURL)
		u, err := url.Parse(configURL)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(u.Scheme)
		fmt.Println(u.User)
		fmt.Println(u.Hostname())
		fmt.Println(u.Port())
	},
}

var (
	configURL string
)

func init() {
	fmt.Println("test init.")
	commandTest.PersistentFlags().StringVarP(&configURL, "config", "u", "", "test url")
	mainCommand.AddCommand(commandTest)
}
