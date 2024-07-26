package commands

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func globalCfgFlag(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("config",
		"c", "/etc/meilibridge/config.yml", "path to config file")
}

func interruptSignal(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()
	return ctx
}
