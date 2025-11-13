package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/osquery/osquery-go"
	"github.com/osquery/osquery-go/plugin/logger"

	"github.com/smithjw/osquery-unified-log-plugin/internal/oslog"
)

func main() {
	socket := flag.String("socket", "/var/osquery/osquery.em", "path to osquery extensions socket")
	_ = flag.Int("timeout", 0, "extension timeout (seconds, passed by osquery)")
	_ = flag.Int("interval", 0, "extension interval (seconds, passed by osquery)")
	name := flag.String("name", "unifiedlog", "logger plugin name")
	structured := flag.Bool("structured", false, "enable structured logging (parse JSON)")
	flag.Parse()

	// Create separate loggers for different categories
	subsystem := "com.github.smithjw.osquery-unified-log-plugin"
	resultsLogger := oslog.New(subsystem, "results")
	snapshotsLogger := oslog.New(subsystem, "snapshots")
	statusLogger := oslog.New(subsystem, "status")

	if resultsLogger == nil || snapshotsLogger == nil || statusLogger == nil {
		log.Fatal("failed to initialize unified log")
	}

	// Logger plugin with optional structured logging support.
	lp := logger.NewPlugin(*name, func(_ context.Context, t logger.LogType, s string) error {
		switch t {
		case logger.LogTypeSnapshot:
			if *structured {
				snapshotsLogger.LogStructured(oslog.TypeDefault, s)
			} else {
				snapshotsLogger.Log(oslog.TypeDefault, s)
			}
		case logger.LogTypeString:
			if *structured {
				resultsLogger.LogStructured(oslog.TypeDefault, s)
			} else {
				resultsLogger.Log(oslog.TypeDefault, s)
			}
		case logger.LogTypeStatus:
			statusLogger.Log(oslog.TypeDefault, s)
		default:
			statusLogger.Log(oslog.TypeDefault, s)
		}
		return nil
	})

	// Minimal server setup, per osquery-go examples.
	server, err := osquery.NewExtensionManagerServer(*name, *socket)
	if err != nil {
		log.Fatalf("extension server: %v", err)
	}
	server.RegisterPlugin(lp)

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutting down gracefully...")
		server.Shutdown(context.Background())
	}()

	if err := server.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
