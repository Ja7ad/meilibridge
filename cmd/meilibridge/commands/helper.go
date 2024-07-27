package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
)

func configFlag(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("config",
		"c", "/etc/meilibridge/config.yml", "path to config file")
}

func interruptSignal(ctx context.Context, log logger.Logger) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		log.WarnContext(ctx, "caught interrupt signal")
		cancel()
	}()
	return ctx
}
