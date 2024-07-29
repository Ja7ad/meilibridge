package commands

import (
	"context"
	"fmt"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/bridge"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
)

func BuildSync(log logger.Logger) *cobra.Command {
	sync := &cobra.Command{
		Use:   "sync",
		Short: "bulk or realtime sync",
	}

	sync.AddCommand(buildBulk(log))

	sync.AddCommand(buildStart(log))

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

		if cfg.General != nil && cfg.General.PProf != nil && cfg.General.PProf.Enable {
			lis := cfg.General.PProf.Listen
			sv := pprofSv(lis)
			log.InfoContext(ctx, "started pprof server",
				"addr", fmt.Sprintf("http://%s/debug/pprof/", lis))
			go func() {
				log.Fatal(sv.ListenAndServe().Error())
			}()
		}

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

	bulk.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := interruptSignal(cmd.Context(), log)

		b, _, err := initBridges(ctx, *cfgPath, log)
		if err != nil {
			return err
		}

		if err := b.BulkSync(ctx, *con); err != nil {
			return err
		}

		return nil
	}

	return bulk
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

	return bridge.New(cfg.Bridges, log), cfg, nil
}
