package main

import (
	"fmt"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/bridge"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"github.com/spf13/cobra"
	"strings"
)

func buildSync() *cobra.Command {
	sync := &cobra.Command{
		Use:   "sync",
		Short: "start realtime sync operation",
	}

	sync.AddCommand(buildBulk())

	sync.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "start realtime sync operation",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	})

	return sync
}

func buildBulk() *cobra.Command {
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

		log := logger.DefaultLogger

		err = database.AddEngine(
			cmd.Context(),
			cfg.Source.Engine,
			cfg.Source.URI,
			cfg.Source.Database,
			log,
		)
		if err != nil {
			return err
		}

		meili, err := meilisearch.New(cmd.Context(), cfg.Meilisearch.APIURL, cfg.Meilisearch.APIKey, log)
		if err != nil {
			return err
		}

		b := bridge.New(cfg.Bridges, meili, cfg.Source.Engine, log)
		if err := b.BulkSync(cmd.Context(), progressBar, *con); err != nil {
			return err
		}

		return nil
	}

	return bulk
}

func progressBar(totalItems, totalIndexedItems int64) {
	percentage := float64(totalIndexedItems) / float64(totalItems) * 100
	barLength := 50
	filledLength := int(float64(barLength) * percentage / 100)
	bar := strings.Repeat("=", filledLength) + strings.Repeat(" ", barLength-filledLength)
	fmt.Printf("\r%.0f%% [%s] (%d/%d)", percentage, bar, totalIndexedItems, totalItems)
}
