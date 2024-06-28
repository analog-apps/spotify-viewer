package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"entgo.io/ent/dialect"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/albe2669/spotify-viewer/ent"
	"github.com/albe2669/spotify-viewer/ent/schema/pulid"
	"github.com/albe2669/spotify-viewer/generated"
	"github.com/albe2669/spotify-viewer/resolver"
	"github.com/albe2669/spotify-viewer/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"

	"github.com/albe2669/spotify-viewer/lib/spotify"
	spotifyLib "github.com/zmb3/spotify/v2"
)

const (
	QueryPath      = "/query"
	PlaygroundPath = "/playground"
	state          = "state" // TODO: unique state string to identify the session, should be random
)

type PlayerState struct {
	Track     ent.Track // This the track struct from the DB schema
	Progress  int       `json:"progress_ms"` // This is the current progress of the track in ms
	Date_Time time.Time // The time that the struct was updated last
}

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

func graphqlServer() *handler.Server {
	server := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{Resolvers: &resolver.Resolver{}},
		),
	)

	return server
}

func getPlayer(sa *spotify.Spotify) echo.HandlerFunc {
	return func(c echo.Context) error {
		player, err := sa.GetPlayerState(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		return c.JSON(http.StatusOK, player)
	}
}

// Global playerstate variable
var playerState *PlayerState

func playerStateLoop(sa *spotify.Spotify, dbClient *ent.Client) {
	// Sleep for 5 seconds to give the server time to start
	time.Sleep(5 * time.Second)

	//TODO: Add a loop restart if error occours
	ctx := context.Background()

	defer dbClient.Close()

	// This is the initial playerstate
	playerState = &PlayerState{}
	for {
		player, err := sa.GetPlayerState(ctx)
		if err != nil {
			// This will more than likely happen in the case where nothing is playing or authentication
			// fails
			utils.Logger.Error("Error getting playerstate", zap.Error(err))
		}

		// Check if player is not nil and that the player has an item
		if player != nil && player.Item != nil {

			// Create a track from the playerstate
			track := makeTrack(player)

			// This function requires data from the previous loop so it needs to be called before the update to the playerstate
			// This is to check if the track has changed and if so add it to the db or if the track has been replayed
			dbCheckUpdate(dbClient, track, player.Progress, ctx)

			// This function updates the playerstate with the new track and progress 
			playerState = updatePlayerState(track, player.Progress)
		}

		// For testing to see if the loop is working
		//utils.Logger.Info("Playerstate recieved", zap.Any("player", playerState))
		// Debugging Query to see if the tracks are being added to the db correctly 
		// Best to use len since it removes some of the clutter from the log
		tr, err := dbClient.Track.Query().All(ctx)
		if err != nil {
			utils.Logger.Error("Error querying tracks", zap.Error(err))
		}
		utils.Logger.Info("Tracks", zap.Any("tracks", len(tr)))

		time.Sleep(5 * time.Second)
	}
}

func makeTrack(player *spotifyLib.PlayerState) *ent.Track {
	// Check if playerState is active if not return previous track as fallback

	artists := make([]string, len(player.Item.Artists))
	//TODO: For some reason the genres field is not available
	// From Item.Artists despite the API saying so. This may be connected
	// to a type issue where Item is not locked to being a track object
	// genres := make(map[string]byte)
	// e.g. player.Item.SimpleTrack.Artists[0].Genres

	for i, artist := range player.Item.Artists {
		artists[i] = artist.Name
	}
	track := &ent.Track{
		Name:          player.Item.Name,
		Artists:       artists,
		ArtistsGenres: nil,
		AlbumName:     player.Item.Album.Name,
		AlbumImageURI: player.Item.Album.Images[0].URL,
		DurationMs:    int32(player.Item.Duration),
		URI:           string(player.Item.URI),
	}

	// Return a pointer to a track
	return track

}

func updatePlayerState(track *ent.Track, progress int) *PlayerState {
	if track.Name == "" {
		return playerState
	}

	//TODO: not sure if this is how you do pointer updates in Golang
	ps := &PlayerState{*track, progress, time.Now()}
	return ps
}

func dbCheckUpdate(dbClient *ent.Client, track *ent.Track, progress int, ctx context.Context) {

	// Check if the track has just changed and if so add it to the db	
	if playerState.Track.Name != track.Name {
		addTrack(dbClient, track, ctx)
	}

	// Check for replays
	// TODO: Maybe find a better way to do this but works for now
	if (int(track.DurationMs/50)*100) < playerState.Progress && progress <= int((track.DurationMs/05)*100) {
		addTrack(dbClient, track, ctx)
	}

}

func addTrack(dbClient *ent.Client, track *ent.Track, ctx context.Context) {
	// NOTE: AritstGenres is not implemented yet so it will be nil see makeTrack function for more info
	_, err := dbClient.Track.Create().
		SetName(track.Name).
		SetArtists(track.Artists).
		SetTrackID(pulid.MustNew(track.Name)).
		SetArtistsGenres(track.ArtistsGenres).
		SetAlbumName(track.AlbumName).
		SetAlbumImageURI(track.AlbumImageURI).
		SetDurationMs(track.DurationMs).
		SetURI(track.URI).
		SetCreatedAt(time.Now()).
		Save(ctx)

	if err != nil {
			utils.Logger.Error("Error creating track", zap.Error(err))
	}
	
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

func connectDatabase() *ent.Client {
	// Create an ent.Client with in-memory SQLite database.
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	ctx := context.Background()
	// Run the automatic migration tool to create all schema resources.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return client
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

	//TODO: Should this be a pointer instead so that we can do reconnects later down the line?
	dbClient := connectDatabase()
	// tracks, _ := dbClient.Track.Query().All(context.Background())
	//log.Println("Tracks:", tracks)

	graphqlServer := graphqlServer()
	e := httpServer(graphqlServer)

	spotify := spotify.NewSpotify(config)
	spotify.SetupRoutes(e)
	spotify.Authenticate()

	go playerStateLoop(spotify, dbClient)
	e.GET("/player", getPlayer(spotify))

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)))
}
