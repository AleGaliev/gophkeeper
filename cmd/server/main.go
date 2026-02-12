package main

import (
	"context"
	cfg "gophkeeper/internal/config/server"
	"gophkeeper/internal/infra/postgres/migrations"
	"gophkeeper/internal/log"
	"gophkeeper/internal/server"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
	serviceName  string = "gophkeeper server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
	defer cancel()

	cfg := cfg.Config{}

	app := &cli.Command{
		Name:  "gophkeeper",
		Usage: "gophkeeper server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "set log level",
				Value:       "INFO",
				Aliases:     []string{},
				Destination: &cfg.LogLevel,
				Sources:     cli.EnvVars("GOPHKEEPER_LOG_LEVEL"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s", "run"},
				Usage:   "start gophkeeper server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "http-port",
						Usage:       "gophkeeper http server port",
						Value:       ":8080",
						Aliases:     []string{"hp"},
						Destination: &cfg.HttpPort,
						Sources:     cli.EnvVars("GOPHKEEPER_HTTP_PORT"),
					},
					&cli.StringFlag{
						Name:        "grpc-port",
						Usage:       "gophkeeper grpc server port",
						Value:       ":8081",
						Aliases:     []string{"gp"},
						Destination: &cfg.GrpcPort,
						Sources:     cli.EnvVars("GOPHKEEPER_GRPC_PORT"),
					},
					&cli.StringFlag{
						Name:        "db-dsn",
						Usage:       "gophkeeper database DSN",
						Value:       "",
						Aliases:     []string{"db"},
						Destination: &cfg.DatabaseDSN,
						Sources:     cli.EnvVars("GOPHKEEPER_DSN"),
					},
					&cli.StringFlag{
						Name:        "key",
						Usage:       "gophkeeper key",
						Value:       "",
						Aliases:     []string{"k"},
						Destination: &cfg.Key,
						Sources:     cli.EnvVars("GOPHKEEPER_KEY"),
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					server, err := server.New(cfg)
					if err != nil {
						return err
					}
					if err = server.Start(ctx); err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "migrate gophkeeper postgres",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "db-dsn",
						Usage:       "gophkeeper database DSN",
						Value:       "",
						Aliases:     []string{"db"},
						Destination: &cfg.DatabaseDSN,
						Sources:     cli.EnvVars("GOPHKEEPER_DN"),
					},
					&cli.BoolFlag{
						Name:        "up",
						Usage:       "migrate up",
						Value:       false,
						Aliases:     []string{"u"},
						Destination: &cfg.Migration.Up,
					},
					&cli.BoolFlag{
						Name:        "down",
						Usage:       "migrate down",
						Value:       false,
						Aliases:     []string{"d"},
						Destination: &cfg.Migration.Down,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {

					psqlMigration, err := migrations.New(cfg.DatabaseDSN, cfg.Logger)
					if err != nil {
						return err
					}
					if cfg.Migration.Up {
						if err = psqlMigration.Up(); err != nil {
							return err
						}
						return nil
					}

					if cfg.Migration.Down {
						if err = psqlMigration.Down(); err != nil {
							return err
						}
						return nil
					}
					cfg.Logger.Info(ctx, "migration not started, select up or down")
					return nil
				},
			},
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			cfg.Logger = log.New(cfg.LogLevel)
			cfg.Logger.Info(ctx, serviceName, "buildVersion", buildVersion, "buildDate", buildDate, "buildCommit", buildCommit)
			return ctx, nil
		},
	}
	if err := app.Run(ctx, os.Args); err != nil {
		cfg.Logger.Error(ctx, err.Error())
		os.Exit(1)
	}

}
