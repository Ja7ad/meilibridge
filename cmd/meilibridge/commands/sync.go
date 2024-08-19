package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/bridge"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
)

func BuildSync(log logger.Logger) *cobra.Command {
	sync := &cobra.Command{
		Use:   "sync",
		Short: "bulk, realtime and trigger sync",
	}

	sync.AddCommand(buildBulk(log))

	sync.AddCommand(buildStart(log))

	sync.AddCommand(buildTrigger(log))

	return sync
}

func buildStart(log logger.Logger) *cobra.Command {
	start := &cobra.Command{
		Use:   "start",
		Short: "start realtime sync operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cfgPath := configFlag(start)

	start.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		b, cfg, err := initBridges(ctx, *cfgPath, log)
		if err != nil {
			return err
		}

		startPProf(log, cfg.General)

		if err := b.Sync(ctx); err != nil {
			return err
		}

		return nil
	}

	return start
}

func buildBulk(log logger.Logger) *cobra.Command {
	bulk := &cobra.Command{
		Use:   "bulk",
		Short: "start bulk sync operation",
	}

	cfgPath := configFlag(bulk)
	con := bulk.Flags().Bool("continue", false, "sync new data on exists index")
	auto := bulk.Flags().Bool("auto", false, "auto bulk sync on exists index every n seconds")

	bulk.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		b, cfg, err := initBridges(ctx, *cfgPath, log)
		if err != nil {
			return err
		}

		if *auto {
			log.Info("auto bulk scheduler started")
			startPProf(log, cfg.General)

			ticker := time.NewTicker(time.Duration(cfg.General.AutoBulkInterval) * time.Second)
			for {
				select {
				case <-cmd.Context().Done():
					ticker.Stop()
					log.Warn("auto bulk sync stopped")
					return nil
				case <-ticker.C:
					if err := b.BulkSync(ctx, true); err != nil {
						return err
					}
				}
			}
		}

		if err := b.BulkSync(ctx, *con); err != nil {
			return err
		}

		return nil
	}

	return bulk
}

func buildTrigger(log logger.Logger) *cobra.Command {
	trigger := &cobra.Command{
		Use:   "trigger",
		Short: "start trigger sync",
	}

	cfgPath := configFlag(trigger)

	trigger.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		b, cfg, err := initBridges(ctx, *cfgPath, log)
		if err != nil {
			return err
		}

		if cfg.General.TriggerSync == nil {
			return errors.New("trigger sync configuration is null")
		}

		startPProf(log, cfg.General)

		return b.TriggerSync(ctx)
	}

	return trigger
}

func initBridges(ctx context.Context, cfgPath string, log logger.Logger) (*bridge.Bridge, *config.Config, error) {
	cfg, err := config.New(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	for _, b := range cfg.Bridges {
		err = database.AddEngine(
			ctx,
			b.Database,
			log,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	return bridge.New(cfg.Bridges, cfg.General.TriggerSync, log), cfg, nil
}

func startPProf(log logger.Logger, general *config.General) {
	if general.PProf != nil && general.PProf.Enable {
		lis := general.PProf.Listen
		sv := pprofSv(lis)
		log.Info("started pprof server",
			"addr", fmt.Sprintf("http://%s/debug/pprof/", lis))
		go func() {
			log.Fatal(sv.ListenAndServe().Error())
		}()
	}
}
