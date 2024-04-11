//+build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/mi-raf/zooad/internal/api"
	database "github.com/mi-raf/zooad/internal/database"
	"github.com/mi-raf/zooad/internal/service"
)

func initApp(ctx context.Context, cfg *config) (a *api.API, closer func(), err error) {
	wire.Build(
		initApiConfig,
		initPostgresConnection,
		database.NewAnimalRepository,
		wire.Bind(new(database.AnimalRepository), new(*database.PgAnimalRepository)),
		service.NewMoodService,
		wire.Bind(new(service.MoodService), new(*service.MoodServiceImpl)),
		service.NewAnimalService,
		api.New,
	)

	return nil, nil, nil
}
