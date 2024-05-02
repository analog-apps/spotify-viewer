package main

import (
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/albe2669/spotify-viewer/generated"
	"github.com/albe2669/spotify-viewer/resolver"
	"github.com/albe2669/spotify-viewer/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/albe2669/spotify-viewer/lib/playerstate"
	spotifyLib "github.com/albe2669/spotify-viewer/lib/spotify"
)

const (
	QueryPath      = "/query"
	PlaygroundPath = "/playground"
	state          = "state" // TODO: unique state string to identify the session, should be random
)

func configureLogger(e *echo.Echo) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			utils.Logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)

			return nil
		},
	}))
}

func graphqlServer(playerStateHandler *playerstate.Handler) *handler.Server {
	server := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{Resolvers: &resolver.Resolver{
				PlayerStateHandler: playerStateHandler,
			}},
		),
	)

	server.AddTransport(&transport.Websocket{})

	return server
}

func httpServer(graphqlServer *handler.Server) *echo.Echo {
	e := echo.New()

	configureLogger(e)

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.POST(QueryPath, echo.WrapHandler(graphqlServer))
	e.GET(PlaygroundPath, func(c echo.Context) error {
		playground.Handler("GraphQL", QueryPath).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return e
}

func main() {
	// Logger
	defer func() {
		_ = utils.Logger.Sync()
	}()

	// Config
	config, err := utils.ReadConfig()
	if err != nil {
		utils.Logger.Fatal("failed reading config", zap.Error(err))
	}

	// playerStateHandler
	playerStateHandler := playerstate.NewHandler()

	graphqlServer := graphqlServer(playerStateHandler)
	e := httpServer(graphqlServer)

	go func() {
		for {
			currTime := new(int)
			*currTime = int(time.Now().UTC().Unix())

			playerStateHandler.Broadcast(&generated.PlayerState{
				Time: currTime,
			})

			time.Sleep(1 * time.Second)
		}
	}()

	spotify := spotifyLib.NewSpotify(config)
	spotify.SetupRoutes(e)
	spotify.Authenticate()

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)))
}
