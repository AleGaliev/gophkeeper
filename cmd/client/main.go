package main

import (
	"context"
	"fmt"
	"gophkeeper/internal/client"
	"gophkeeper/internal/config/agent"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
	serviceName  string = "gophkeeper client"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
	defer cancel()

	cfg := agent.Config{}

	app := &cli.Command{
		Name:  "gophkeeper",
		Usage: "gophkeeper client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "grpc server address",
				Value:       "",
				Aliases:     []string{"a"},
				Destination: &cfg.GrpcAddr,
				Sources:     cli.EnvVars("GOPHKEEPER_GRPC_ADDR"),
			},
			&cli.StringFlag{
				Name:        "user",
				Usage:       "username",
				Value:       "",
				Aliases:     []string{"u"},
				Destination: &cfg.User.Login,
				Sources:     cli.EnvVars("GOPHKEEPER_AUTH_USER"),
			},
			&cli.StringFlag{
				Name:        "password",
				Usage:       "password",
				Value:       "",
				Aliases:     []string{"p"},
				Destination: &cfg.User.Password,
				Sources:     cli.EnvVars("GOPHKEEPER_AUTH_PASSWORD"),
			},
			&cli.StringFlag{
				Name:        "key-path",
				Usage:       "rsa private key",
				Value:       "~/.gophkeeper/id_rsa",
				Aliases:     []string{"i"},
				Destination: &cfg.KeyPath,
				Sources:     cli.EnvVars("GOPHKEEPER_PRIVATE_KEY"),
			},
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			fmt.Printf("Build version: %s, Build date: %s, Build commit: %s\n", buildVersion, buildDate, buildCommit)
			return ctx, nil
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			c, err := client.New(cfg)
			if err != nil {
				return err
			}

			if err := c.Start(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
