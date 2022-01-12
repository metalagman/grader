package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"grader/internal/app/queue/app"
	"grader/pkg/logger"
	"os"
	"os/signal"
)

// serveCmd represents the migrate command
var serveCmd = &cobra.Command{
	Use:   "work",
	Short: "Start queue",
	Run: func(cmd *cobra.Command, args []string) {
		work()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func work() {
	l := logger.Global()

	// setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		osCall := <-stop
		logger.Global().Info().Str("signal", fmt.Sprintf("%+v", osCall)).Msg("System call")
		cancel()
	}()

	a, err := app.New(cfg)
	if err != nil {
		logger.CheckErr(err)
	}

	go func() {
		<-ctx.Done()
		a.Stop()
	}()

	l.Info().Msg("Application started")

	if err := a.Sender.Send(); err != nil {
		logger.CheckErr(err)
	}

	l.Info().Msg("Application stopped")
}
