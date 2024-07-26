package main

import (
	"context"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	root := &cobra.Command{
		Use:   "meilibridge",
		Short: "Meilibridge",
		Long: "Meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL " +
			"databases to Meilisearch, providing an efficient and unified search solution.",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	log := logger.DefaultLogger

	root.AddCommand(buildSync(log))

	err := root.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}

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
