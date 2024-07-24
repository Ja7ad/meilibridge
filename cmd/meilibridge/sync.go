package main

import (
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
)

func buildSync(logger logger.Logger, cfgPath string) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "start realtime sync operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New(cfgPath)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
