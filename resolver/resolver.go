package resolver

import "github.com/albe2669/spotify-viewer/lib/playerstate"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	PlayerStateHandler *playerstate.Handler
}
