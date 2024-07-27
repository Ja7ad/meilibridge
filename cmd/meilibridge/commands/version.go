package commands

import (
	"fmt"

	"github.com/Ja7ad/meilibridge/version"
	"github.com/spf13/cobra"
)

func BuildVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Version())
		},
	}
}
