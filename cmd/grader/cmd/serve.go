package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"grader/internal/app/grader/app"
	"grader/pkg/logger"
	"os"
	"os/signal"
)

// serveCmd represents the migrate command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving grader requests",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("listen", "l", "localhost:8090", "server address and port to listen on")
	logger.CheckErr(viper.BindPFlag("server.listen", serveCmd.Flags().Lookup("listen")))
}

func serve() {
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

	l.Info().Msg("Application started")

	<-ctx.Done()

	a.Stop()
	l.Info().Msg("Application stopped")
}
