package main

import (
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/bridge"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"github.com/spf13/cobra"
)

func buildSync(log logger.Logger) *cobra.Command {
	sync := &cobra.Command{
		Use:   "sync",
		Short: "bulk or realtime sync",
	}

	sync.AddCommand(buildBulk(log))

	sync.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "start realtime sync operation",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	})

	return sync
}

func buildBulk(log logger.Logger) *cobra.Command {
	bulk := &cobra.Command{
		Use:   "bulk",
		Short: "start bulk sync operation",
	}

	cfgPath := globalCfgFlag(bulk)
	con := bulk.Flags().Bool("continue", false, "sync new data on exists index")

	bulk.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New(*cfgPath)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		for _, b := range cfg.Bridges {
			err = database.AddEngine(
				cmd.Context(),
				b.Source.Engine,
				b.Source.URI,
				b.Source.Database,
				log,
			)
			if err != nil {
				return err
			}
		}

		meili, err := meilisearch.New(cmd.Context(), cfg.Meilisearch.APIURL, cfg.Meilisearch.APIKey, log)
		if err != nil {
			return err
		}

		b := bridge.New(cfg.Bridges, meili, log)
		if err := b.BulkSync(cmd.Context(), *con); err != nil {
			return err
		}

		return nil
	}

	return bulk
}
