package playerstate

import (
	"context"
	"sync"

	"github.com/albe2669/spotify-viewer/generated"
	"github.com/google/uuid"
)

type Handler struct {
	clients map[string]chan *generated.PlayerState
	mu      sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		clients: make(map[string]chan *generated.PlayerState),
	}
}

func (h *Handler) AddClient(ctx context.Context) (clientID string, ch chan *generated.PlayerState) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clientID = uuid.New().String()

	ch = make(chan *generated.PlayerState, 1)
	h.clients[clientID] = ch

	go func() {
		<-ctx.Done()
		h.RemoveClient(clientID)
	}()

	return clientID, ch
}

func (h *Handler) RemoveClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, clientID)
}

func (h *Handler) Broadcast(playerState *generated.PlayerState) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range h.clients {
		ch <- playerState
	}
}
