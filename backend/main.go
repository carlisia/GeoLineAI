// Command geolineai serves GeoLineAI, or with a subcommand imports a survey CSV or
// resolves the sample scenarios against the database.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"geolineai/internal/cli"
	"geolineai/internal/core/assist/prompts"
	"geolineai/internal/core/intake/importer"
	"geolineai/internal/core/scenarios"
	"geolineai/internal/integrations/db"
	"geolineai/internal/integrations/openrouter"
	"geolineai/internal/server"
)

// Config is read from the environment, each variable prefixed GEOLINEAI_. Relative
// paths resolve from the working directory, which the scripts set to the repository
// root.
type Config struct {
	DBPath   string `envconfig:"DB_PATH" default:"var/geolineai.db"`
	HTTPPort int    `envconfig:"HTTP_PORT" default:"18080"`
	// OpenRouterAPIKey reaches both models; empty leaves explanations unavailable and
	// changes nothing else.
	OpenRouterAPIKey string `envconfig:"OPENROUTER_API_KEY"`
}

// The files the server reads beside the database, relative to the repository root.
const (
	webRoot      = "frontend/dist"
	manifestPath = "data/scenarios.json"
)

const usage = `usage: geolineai [command]

Commands:
	(none)      serve the app on 127.0.0.1:GEOLINEAI_HTTP_PORT (default 18080)
	import      load a survey CSV into the database at GEOLINEAI_DB_PATH
	scenarios   resolve data/scenarios.json against the database and print it
`

func main() {
	var conf Config
	if err := envconfig.Process("GEOLINEAI", &conf); err != nil {
		fmt.Fprintf(os.Stderr, "read configuration: %v\n", err)
		os.Exit(2)
	}
	ctx := context.Background()
	if len(os.Args) > 1 {
		switch cmd := os.Args[1]; cmd {
		case "import":
			open := func(ctx context.Context) (importer.Committer, func() error, error) {
				d, err := db.Open(ctx, conf.DBPath)
				return d, d.Close, err
			}
			os.Exit(cli.Import(ctx, os.Args[2:], open, os.Stdout, os.Stderr))
		case "scenarios":
			os.Exit(printScenarios(ctx, conf))
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
			os.Exit(2)
		}
	}
	os.Exit(serve(ctx, conf))
}

// printScenarios resolves the manifest and prints each scenario's reading ID, or
// names the scenario that does not resolve.
func printScenarios(ctx context.Context, conf Config) int {
	entries, err := scenarios.Load(manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	d, err := db.Open(ctx, conf.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open database %s: %v\n", conf.DBPath, err)
		return 1
	}
	defer func() { _ = d.Close() }()
	list, err := scenarios.Resolve(ctx, d, entries)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, s := range list.Scenarios {
		fmt.Printf("%s: reading %s\n", s.Label, s.ReadingID)
	}
	return 0
}

func serve(ctx context.Context, conf Config) int {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	entries, err := scenarios.Load(manifestPath)
	if err != nil {
		slog.Error("failed to read the scenario manifest", slog.String("error", err.Error()))
		return 1
	}
	d, err := db.Open(ctx, conf.DBPath)
	if err != nil {
		slog.Error("failed to open database", slog.String("path", conf.DBPath), slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := d.Close(); err != nil {
			slog.Warn("failed to close database cleanly", slog.String("error", err.Error()))
		}
	}()

	var models server.Models
	if conf.OpenRouterAPIKey == "" {
		slog.Warn("model help is off: GEOLINEAI_OPENROUTER_API_KEY is not set")
	} else {
		client := openrouter.New(conf.OpenRouterAPIKey, openrouter.DefaultEndpoints)
		models = server.Models{Writer: client, Judge: client}
	}
	routes := server.Routes(d, models, server.Config{Manifest: entries, WebRoot: webRoot, Fingerprint: prompts.Fingerprint})

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(conf.HTTPPort))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("cannot listen; is the port in use? set GEOLINEAI_HTTP_PORT to another",
			slog.String("address", addr), slog.String("error", err.Error()))
		return 1
	}
	srv := &http.Server{Handler: routes, ReadHeaderTimeout: 10 * time.Second}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()
	slog.Info("serving GeoLineAI", slog.String("url", "http://"+addr), slog.String("database", conf.DBPath))

	select {
	case err := <-served:
		slog.Error("server stopped", slog.String("error", err.Error()))
		return 1
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to shut down cleanly", slog.String("error", err.Error()))
		return 1
	}
	return 0
}
