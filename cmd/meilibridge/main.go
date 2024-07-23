package main

import (
	"github.com/spf13/cobra"
	"log"
)

func main() {
	root := &cobra.Command{
		Use:   "meilibridge",
		Short: "Meilibridge",
		Long: "Meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL " +
			"databases to Meilisearch, providing an efficient and unified search solution.",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	buildSync(root)
	buildBulk(root)

	err := root.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
