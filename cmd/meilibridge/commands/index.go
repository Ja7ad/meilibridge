package commands

import (
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"github.com/spf13/cobra"
)

func BuildIndex(log logger.Logger) *cobra.Command {
	index := &cobra.Command{
		Use:   "index",
		Short: "manage index",
	}

	index.AddCommand(buildIndexSettingsUpdate(log))
	index.AddCommand(buildCreateIndex(log))

	return index
}

func buildCreateIndex(log logger.Logger) *cobra.Command {
	create := &cobra.Command{
		Use:   "create",
		Short: "create index with update settings",
	}

	cfgPath := configFlag(create)

	create.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		log.Info("loading configuration...")

		cfg, err := config.New(*cfgPath)
		if err != nil {
			return err
		}

		for _, bridge := range cfg.Bridges {
			log.Info("started creating index", "bridge", bridge.Name)

			if bridge.Meilisearch == nil {
				log.Warn("not available meilisearch configuration", "bridge", bridge.Name)
				continue
			}

			log.Info("connecting to meilisearch", "bridge", bridge.Name)

			meili, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, log)
			if err != nil {
				log.Warn("failed to create meilisearch client", "error", err)
				continue
			}

			for _, idx := range bridge.IndexMap {
				log.Info("creating index", "index", idx.IndexName)
				if err := meili.CreateIndex(ctx, idx.IndexName, idx.PrimaryKey); err != nil {
					log.Warn("failed to create meilisearch index", "index", idx.IndexName, "error", err)
					continue
				}

				log.Info("updating index settings", "index", idx.IndexName)
				if err := meili.UpdateIndexSettings(ctx, idx.IndexName, idx.Settings); err != nil {
					log.Warn("failed to update index settings", "index", idx.IndexName, "error", err)
				}
			}

			log.Info("completed index update", "bridge", bridge.Name)
		}

		return nil
	}

	return create
}

func buildIndexSettingsUpdate(log logger.Logger) *cobra.Command {
	settings := &cobra.Command{
		Use:   "settings",
		Short: "manage settings",
	}

	update := &cobra.Command{
		Use:   "update",
		Short: "update settings",
	}

	settings.AddCommand(update)

	cfgPath := configFlag(update)

	update.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		log.Info("loading configuration...")

		cfg, err := config.New(*cfgPath)
		if err != nil {
			return err
		}

		for _, bridge := range cfg.Bridges {

			log.Info("started update index", "bridge", bridge.Name)

			if bridge.Meilisearch == nil {
				log.Warn("not available meilisearch configuration", "bridge", bridge.Name)
				continue
			}

			log.Info("connecting to meilisearch", "bridge", bridge.Name)

			meili, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, log)
			if err != nil {
				log.Warn("failed to create meilisearch client", "error", err)
				continue
			}

			for _, idx := range bridge.IndexMap {
				log.Info("updating index settings", "index", idx.IndexName)
				if err := meili.UpdateIndexSettings(ctx, idx.IndexName, idx.Settings); err != nil {
					log.Warn("failed to update index settings", "index", idx.IndexName, "error", err)
				}
			}

			log.Info("completed update index", "bridge", bridge.Name)
		}

		return nil
	}

	return settings
}
