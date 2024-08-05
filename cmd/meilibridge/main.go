package main

import (
	"fmt"

	"github.com/Ja7ad/meilibridge/cmd/meilibridge/commands"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/version"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "meilibridge",
		Short: fmt.Sprintf("Meilibridge %s", version.Version()),
		Long: "Meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL " +
			"databases to Meilisearch, providing an efficient and unified search solution.",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	log := logger.DefaultLogger

	root.AddCommand(commands.BuildSync(log))
	root.AddCommand(commands.BuildVersion())
	root.AddCommand(commands.BuildIndex(log))

	err := root.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
