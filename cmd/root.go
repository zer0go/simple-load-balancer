package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
	"github.com/zer0go/simple-load-balancer/internal/config"
	"github.com/zer0go/simple-load-balancer/internal/service"
)

var rootCmd = &cobra.Command{
	Use:               "lb",
	Short:             "Simple Load Balancer",
	SilenceErrors:     true,
	PersistentPreRunE: bootstrap,
	RunE:              run,
}

func init() {
	rootCmd.
		PersistentFlags().
		CountP("verbosity", "v", "set logging verbosity")
}

func bootstrap(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetCount("verbosity")

	logLevel := slog.LevelInfo
	if verbosity > 0 {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   true,
		Level:       logLevel,
		ReplaceAttr: nil,
	})))

	if err := config.LoadAppConfig(); err != nil {
		return err
	}
	slog.Debug("config loaded", "config", config.Get())

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	lb := service.NewLoadBalancer(config.Get().BackendUrls, config.Get().HealthCheckPath)
	lb.StartHealthChecks(time.Duration(config.Get().HealthCheckIntervalSeconds) * time.Second)

	slog.Info("service started",
		"url", fmt.Sprintf("http://%s", config.Get().Address),
	)
	return http.ListenAndServe(config.Get().Address, lb)
}

func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Short += " " + version

	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered", "panic", r, "stack", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
