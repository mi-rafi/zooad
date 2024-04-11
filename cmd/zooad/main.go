package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mi-raf/zooad/internal/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

func main() {
	defer closer.Close()

	closer.Bind(func() {
		log.Info().Msg("shutdown")
	})

	cfg, err := initConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't init config")
	}

	if err := initLogger(cfg); err != nil {
		log.Fatal().Err(err).Msg("Can't init logger")
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	closer.Bind(cancelCtx)

	a, cleanup, err := initApp(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't init app")
	}
	closer.Bind(cleanup)
	closer.Bind(func() {
		if err := a.Close(); err != nil {
			log.Error().Err(err).Msg("Can't stop web application")
		}
	})
	if err := a.Start(); err != nil {
		log.Fatal().Err(err).Msg("Can't start app")
	}
}

func initLogger(c *config) error {
	log.Debug().Msg("init logger")
	logLvl, err := zerolog.ParseLevel(strings.ToLower(c.LogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)
	switch c.LogFmt {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
	default:
		return fmt.Errorf("unknown output format %s", c.LogFmt)

	}
	return nil
}

func initPostgresConnection(ctx context.Context, cfg *config) (*pgxpool.Pool, func(), error) {
	//todo write New rep
	pg, err := pgxpool.New(ctx, cfg.DbAddr)
	if err != nil {
		return nil, nil, err
	}
	err = pg.Ping(ctx)
	if err != nil {
		return nil, nil, err
	}

	return pg, pg.Close, nil
}

func initApiConfig(cfg *config) *api.Config {
	return &api.Config{Addr: cfg.Listen}
}
