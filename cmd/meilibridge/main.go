package main

import (
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "meilibridge",
		Short: "Meilibridge",
		Long: "Meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL " +
			"databases to Meilisearch, providing an efficient and unified search solution.",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	cfgPath := root.Flags().StringP("config", "c", "./config.yml", "path to config file")

	log := logger.DefaultLogger

	root.AddCommand(buildSync(log, *cfgPath))
	root.AddCommand(buildBulk(log, *cfgPath))

	err := root.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
